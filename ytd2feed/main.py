import app
import os
import sys
import time
from multiprocessing import Process
from importlib.metadata import version

def start_download(fc: app.FeedCreater,
                   format: str,
                   retries: int,
                   ytdl_link: str,
                   taskId: int,
                   m_client: app.McoreClient):
    
    print("try to download link:", ytdl_link)
    msg = "downloaded"
    try:
        app.download(ytdl_link, format, retries, fc)
    except:
        msg = "some error in downloaded"

    success = m_client.report_task(taskId, 4, msg)
    if success:
        print("Download complited")
    else:
        print("Error")



if __name__ == '__main__':
    print("start app")
    print(version('yt_dlp'))
    sys.exit()
    cfg = app.Conf()
    print("mcore addr: {}, mcore_secret: {}, taskType: {}, "
          .format(cfg.mcore_addr, cfg.mcore_secret,cfg.task_type))
    # init content path
    if not os.path.isdir(cfg.path2content):
         os.makedirs(cfg.path2content)

    print("init mcore client")
    m_client = app.McoreClient(cfg.mcore_addr, cfg.task_type, cfg.mcore_secret)
    if not m_client.health():
        print("Mcore health failed")
        sys.exit()
    print("Start to listen")
    while True:
        d = m_client.get_task()
        if d.get("id") is None:
            time.sleep(1)  # be nice to the system :)
            continue
        if m_client.health_reported(d):
            continue
        if not m_client.check_and_report(d):
            continue
        print("start to process taskid:",str(d["id"]), "link:", d["taskData"]["ytdl"]["link"])
        fo = cfg.get_user_conf(d["taskData"]["ytdl"]["userName"])
        fc = app.FeedCreater(cfg.path2content, cfg.url2content, fo["feedName"], fo["feedDescription"])
        process = Process(target=start_download, args=(fc,
                                                       fo["format"],
                                                       cfg.retries,
                                                       d["taskData"]["ytdl"]["link"],
                                                       d["id"],
                                                       m_client))
        process.start()

