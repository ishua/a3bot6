import yt_dlp
import os
from app import FeedCreater


class PPadd2feed(yt_dlp.postprocessor.PostProcessor):

    def run(self, info):

        self.fc.add2rss(
            title = info['title'], 
            description = info["description"],
            webpageUrl = info['webpage_url'],
            fileExtention = info['ext'] ,
            fileSize = str(info.get('filesize', ""))
        )
        return [], info


def download(url: str, format: str, retries: int, fc: FeedCreater):

    filename = fc.getFileName()
    filepath = str(os.path.join(fc.getPath2content(),
                  fc.getFeedName(), filename)) + '.%(ext)s'


    ydl_opts = {
        'outtmpl': filepath,
        'format': format,
        'retries': retries,
        'sleep-interval': 10,
        'http_headers': {'Referer': 'https://www.google.com'}
    }

    with yt_dlp.YoutubeDL(ydl_opts) as ydl:
        ppadd2feed = PPadd2feed()
        ppadd2feed.fc = fc
        ydl.add_post_processor(ppadd2feed)

        ydl.download(url)

