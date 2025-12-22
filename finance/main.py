#!/usr/bin/python3
import os
import sys
from multiprocessing import Process
import time
import app

def go_command(mclient: app.McoreClient,
               task_id: int,
               command: str
               ):
    
    print("go_command task_id: {}, command: {}"
          .format(task_id, command))
    
    msg = "done"


    mclient.report_task(task_id, 4, msg)


if __name__ == '__main__':
    print("start app")
    cfg = app.Conf()
    print("mcore addr: {}, mcore_secret: {}, taskType: {}"
          .format(cfg.mcore_addr, cfg.mcore_secret,cfg.task_type))

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

        print("start to process taskid:",str(d["id"]), "command:", d["taskData"]["fin"]["command"])
        process = Process(target=go_command, args=(
            mclient,
            d["id"],
            d["taskData"]["fin"]["command"]
        ))
        process.start()