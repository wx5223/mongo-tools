#!/usr/bin/env python

from __future__ import print_function

import collections
import copy
import datetime
import os
import sys
import time

from optparse import OptionParser

import psutil


def get_all_processes():
    all_processes = []
    for pid in psutil.pids():
        try:
            pname = psutil.Process(pid).name()
            all_processes.append((pid, pname))
        except psutil.NoSuchProcess:
            pass
    return all_processes


def pname_match(match_type, pname, processes):
    pname = os.path.splitext(pname)[0]
    for ip in processes:
        if (match_type == "exact" and pname == ip or match_type == "contains" and ip in pname):
            return True
    return False


def get_proc_info(pid):
    p = psutil.Process(pid)
    with p.oneshot():
        proc_info = p.as_dict()
    return proc_info


def get_mount_points():
    mount_points = []
    for disk in psutil.disk_partitions():
        mount_points.append((getattr(disk, "device", None), getattr(disk, "mountpoint", None)))
    return mount_points


def get_disks():
    disk_info = psutil.disk_io_counters(perdisk=True)
    return disk_info.keys()


def disk_usage(device, disk_name):
    return (disk_name, psutil.disk_usage(disk_name), psutil.disk_usage(disk_name))


def get_cpu_info(proc_info):
    cpu_info = collections.OrderedDict({"cpu_percent": proc_info["cpu_percent"]})
    proc_cpu = proc_info.get("cpu_times")
    if proc_cpu:
        for field in proc_cpu._fields:
            cpu_info[field] = getattr(proc_cpu, field)
    return cpu_info


def get_mem_info(proc_info):
    mem_info = collections.OrderedDict({"memory_percent": proc_info["memory_percent"]})
    proc_mem = proc_info.get("memory_full_info")
    if not proc_mem:
        proc_mem = proc_info.get("memory_info")
    if proc_mem:
        for field in proc_mem._fields:
            mem_info[field] = getattr(proc_mem, field)
    return mem_info


def get_io_info(proc_info):
    io_info = collections.OrderedDict({"num_open_files": len(proc_info["open_files"])})
    proc_io = proc_info.get("io_counters")
    if proc_io:
        for field in proc_io._fields:
            io_info[field] = getattr(proc_io, field)
    return io_info


def get_system_io_info(disk, last_system_io_info, interval):
    system_io_info = collections.OrderedDict()
    disk_io_counters = psutil.disk_io_counters(perdisk=True)
    if disk not in disk_io_counters:
        return system_io_info
    disk_info = disk_io_counters[disk]
    for field in disk_info._fields:
        system_io_info[field] = getattr(disk_info, field)
        if field in ["read_bytes", "write_bytes"]:
            byte_rate_field = "{}_per_sec".format(field)
            if field in last_system_io_info:
                byte_rate = ((getattr(disk_info, field) - last_system_io_info[field]) /
                             interval)
            else:
                byte_rate = 0
            system_io_info[byte_rate_field] = byte_rate
    return system_io_info


def get_other_info(proc_info):
    other_info = collections.OrderedDict({"num_threads": proc_info["num_threads"]})
    return other_info


def compute_io_stats(system_io_info, interval):
    computed_system_io_info = copy.deepcopy(system_io_info)
    return computed_system_io_info


def main():

    parser = OptionParser(description=__doc__)

    parser.add_option("-m", "--match",
                      dest="process_match",
                      choices=["contains", "exact"],
                      default="contains",
                      help="Type of match for process names (-p & -g), specify 'contains', or"
                           " 'exact'. Note that the process name match performs the following"
                           " conversions: change all process names to lowecase, strip off the file"
                           " extension, like '.exe' on Windows. Default is 'contains'.")
    parser.add_option("-p", "--names",
                      dest="process_names",
                      help="Comma separated list of process names to analyze.")
    parser.add_option("-d", "--pids",
                      dest="process_ids",
                      default=None,
                      help="Comma separated list of process ids (PID) to analyze, overrides -p &"
                           " -g.")
    parser.add_option("-i", "--interval",
                      dest="interval",
                      default=30,
                      type=int,
                      help="Interval frequency to report results."
                           "Defaults to '%default'.")
    parser.add_option("--procFile",
                      dest="proc_file",
                      default="proc.log",
                      help="Output file for Process statistics. Defaults to '%default'.")
    parser.add_option("--ioFile",
                      dest="io_file",
                      default="io.log",
                      help="Output file for I/O statistics. Defaults to '%default'.")

    parser.add_option("-f", "--fork",
                      dest="fork",
                      action="store_true",
                      default=False,
                      help="Fork this process to run in the background.")

    (options, args) = parser.parse_args()

    print("Saving process information to {}.".format(options.proc_file))
    print("Saving I/O information to {}.".format(options.io_file))
    if options.fork:
        pr, pw = os.pipe()
        child_pid = os.fork()
        if child_pid != 0:
            print("Forked child process {}, exiting parent.".format(child_pid))
            os._exit(0)
        os.dup2(pw, sys.stdout.fileno())

    proc_fh = open(options.proc_file, "w+")
    io_fh = open(options.io_file, "w+")

    io_header_printed = False
    proc_header_printed = False
    last_system_io_info = {}
    while True:
        process_ids = []
        interesting_processes = []
        if options.process_ids is not None:
            # process_ids is an int list of PIDs
            process_ids = [int(pid) for pid in options.process_ids.split(',')]

        if options.process_names is not None:
            interesting_processes = options.process_names.split(',')

        processes = []
        all_processes = [(pid, pname.lower()) for (pid, pname) in get_all_processes()]

        # Find all running interesting processes:
        #   If a list of process_ids is supplied, match on that.
        #   Otherwise, do a substring match on interesting_processes.
        if process_ids:
            processes = [(pid, pname) for (pid, pname) in all_processes
                         if pid in process_ids and pid != os.getpid()]
            running_pids = set([pid for (pid, pname) in all_processes])
            missing_pids = set(process_ids) - running_pids
            if missing_pids:
                print("The following requested process ids are not running {}."
                      .format(list(missing_pids)))

        else:
            processes = [(pid, pname) for (pid, pname) in all_processes
                         if pname_match(options.process_match, pname, interesting_processes) and
                         pid != os.getpid()]

        timestamp = datetime.datetime.utcnow().isoformat()
        print("{} Found {:d} interesting process(es) {}".format(
            timestamp, len(processes), processes))

        for disk in get_disks():
            timestamp = datetime.datetime.utcnow().isoformat()
            system_io_info = get_system_io_info(disk,
                                                last_system_io_info.get(disk, {}),
                                                options.interval)
            if not io_header_printed:
                io_fh.write("time disk {}\n".format(" ".join(system_io_info.keys())))
                io_header_printed = True
            last_system_io_info[disk] = copy.deepcopy(system_io_info)
            io_fh.write("{} {} {}\n".format(
                timestamp,
                disk,
                " ".join(str(x) for x in system_io_info.values())))
            io_fh.flush()

        for pid, pname in processes:
            proc_info = get_proc_info(pid)
            timestamp = datetime.datetime.utcnow().isoformat()
            cpu_info = get_cpu_info(proc_info)
            mem_info = get_mem_info(proc_info)
            io_info = get_io_info(proc_info)
            other_info = get_other_info(proc_info)
            if not proc_header_printed:
                proc_fh.write("time name pid {} {} {} {}\n".format(
                    " ".join(cpu_info.keys()),
                    " ".join(mem_info.keys()),
                    " ".join(io_info.keys()),
                    " ".join(other_info.keys())))
                proc_header_printed = True
            proc_fh.write("{} {} {} {} {} {} {}\n".format(
                timestamp,
                pname,
                pid,
                " ".join(str(x) for x in cpu_info.values()),
                " ".join(str(x) for x in mem_info.values()),
                " ".join(str(x) for x in io_info.values()),
                " ".join(str(x) for x in other_info.values())))
            proc_fh.flush()
        time.sleep(options.interval)

if __name__ == "__main__":
    main()
