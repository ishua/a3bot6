#!/usr/bin/python3
import app
import os
import redis
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
    print("redis host: {}, listern channel: {}".format(cfg.redis, cfg.channel))
    # init content path
    if not os.path.isdir(cfg.path2content):
         os.makedirs(cfg.path2content)

    # init redis
    r = redis.Redis(host=cfg.redis)
    p = r.pubsub()
    p.subscribe(cfg.channel)

    print("Start to lisen")
    while True:
        try:
            message = p.get_message()
        except redis.ConnectionError:
            # Do reconnection attempts here such as sleeping and retrying
            print("reconnect to redis after 3 sec")
            time.sleep(3)
            p = r.pubsub()
            p.subscribe(cfg.channel)
        if message:
            if message["type"] == "message":
                payload = message["data"]
                a = app.Answer(payload, cfg.redis, cfg.tbotchannel)
                fo = cfg.getUserCnf(a.getUserName())
                if fo is None:
                    a.setReply("config not found, user: " + a.getUserName())
                # start download in background
                proccess = Process(target=start_download, args=(cfg.path2content, 
                                                                cfg.url2content, 
                                                                fo["feedName"], 
                                                                fo["feedDescription"], 
                                                                fo["format"], 
                                                                fo["retries"],
                                                                a,))
                proccess.start()
        time.sleep(1)  # be nice to the system :)


