import app
import os
from multiprocessing import Process
import time

def start_download(path2content: str,
                url2content: str,
                feedName: str,
                feedDescription: str,
                format: str,
                retries: int,
                answerer: app.Answer):
    
    print("try to download")
    fc = app.FeedCreater(path2content, url2content, feedName, feedDescription)
    err = answerer.validateError()
    msg = "downloaded"
    if err:
        print("err:", answerer.getReply())
        answerer.send()
        return
    try:
        app.download(answerer.getUrl(), format, retries, fc)
    except:
        msg = "some error"

    answerer.setReply(msg)
    answerer.send()
    print("Download complited")



if __name__ == '__main__':
    print("start app")
    cfg = app.Conf()
    print("mcore addr: {}, taskType: {}".format(cfg.mcore_addr, cfg.task_type))
    # init content path
    if not os.path.isdir(cfg.path2content):
         os.makedirs(cfg.path2content)


    print("Start to lisen")
