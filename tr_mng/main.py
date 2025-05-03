#!/usr/bin/python3
import app
import os
import sys
from multiprocessing import Process
import time

def go_command(mclient: app.McoreClient,
               trhost: str,
               trport: int,
               tdownloaddir: str,
               task_id: int,
               command: str,
               torrent_url: str,
               folder_path: str,
               torrent_id: int):
    print("go_command trhost: {}, trport: {}, tdownloaddir: {}, task_id: {}, command: {}, torrent_url: {}, folder_path: {}, torrent_id: {}"
          .format(trhost, trport, tdownloaddir, task_id, command, torrent_url, folder_path, torrent_id))
    t = app.Tr(trhost, trport, tdownloaddir)
    msg = ""
    if command == "add":
        msg = t.add_torrent(torrent_url, folder_path)
    if command == "list":
        msg = t.list_torrents()
    if command == "del":
        msg = t.del_torrent(torrent_id)

    mclient.report_task(task_id, 4, msg)


if __name__ == '__main__':
    print("start app")
    cfg = app.Conf()
    print("mcore addr: {}, mcore_secret: {}, taskType: {}, trhost: {}, trport: {}, tdownloaddir: {}"
          .format(cfg.mcore_addr, cfg.mcore_secret,cfg.task_type, cfg.trhost, cfg.trport, cfg.tdownloaddir))

    mclient = app.McoreClient(cfg.mcore_addr, cfg.task_type, cfg.mcore_secret)
    mclient.health()
    print("Start to listen")
    while True:
        d = mclient.get_task()
        if d.get("id") is None:
            time.sleep(1)  # be nice to the system :)
            continue
        if mclient.health_reported(d):
            continue
        if not mclient.check_and_report(d):
            continue

        print("start to process taskid:",str(d["id"]), "command:", d["taskData"]["tr"]["command"])
        process = Process(target=go_command, args=(
            mclient,
            cfg.trhost,
            cfg.trport,
            cfg.tdownloaddir,
            d["id"],
            d["taskData"]["tr"]["command"],
            d["taskData"]["tr"]["torrentUrl"],
            d["taskData"]["tr"]["folderPath"],
            d["taskData"]["tr"]["torrentId"]
        ))
        process.start()