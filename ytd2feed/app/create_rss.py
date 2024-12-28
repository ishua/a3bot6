
import os.path
import xml.etree.ElementTree as ET
from urllib.parse import urljoin
from datetime import datetime
import string
import random
from datetime import date


class FeedCreater:
    def __init__( self, path2content: str, 
                 url2content: str,
                 feedName: str, 
                 feedDescription: str 
                 ):
        self.feedName = feedName
        self.url2content = url2content
        self.feedDescription = feedDescription
        self.path2content = path2content
        self.path2rss = os.path.join(path2content, feedName + ".xml")
        self.filename = ""

    def _feed_not_exist(self) -> bool:
        if not os.path.isdir(self.path2content):
            os.makedirs(self.path2content)
            return True

        if not os.path.isfile(self.path2rss):
            return True
        return False

    def _create_new_feed(self):

        rss = ET.Element("rss",  version="2.0")
        channel = ET.SubElement(rss, "channel")
        ET.SubElement(
            channel, "title").text = self.feedName
        ET.SubElement(channel, "link").text = urljoin(
            self.url2content, self.feedName + ".xml")
        ET.SubElement(channel, "language").text = 'ru'
        ET.SubElement(channel, "copyright").text = 'BSD'
        # ET.SubElement(channel, "author").text = i['a3b_info']['rss_email']
        ET.SubElement(
            channel, "description").text = self.feedDescription
        # ET.SubElement(channel, "thumbnail").text = i['a3b_info']['rss_thumbnail']
        # ET.SubElement(channel, "credit", role="author").text = i['a3b_info']['user']
        ET.SubElement(channel, "rating").text = 'nonadult'

        tree = ET.ElementTree(rss)
        ET.indent(tree, space="\t", level=0)

        tree.write(self.path2rss, xml_declaration=True, encoding="utf-8", )
    
    def getFileName(self) -> str:
        if self.filename == "":
            self.filename = date.today().strftime("%y%m%d")
            self.filename = self.filename + "".join(random.choice(string.ascii_lowercase)
                            for _ in range(6))
        
        return self.filename
    
    def getFeedName(self) -> str:
        return self.feedName

    def getPath2content(self) -> str:
        return self.path2content

    def add2rss(self, 
                title: str,  
                description: str,
                webpageUrl: str,
                fileExtention: str,
                fileSize: str):
        if self._feed_not_exist():
            self._create_new_feed()
        
        parser = ET.XMLParser(encoding="utf-8")
        tree = ET.parse(self.path2rss, parser=parser)
        root = tree.getroot()
        channel = root.find("channel")

        item = ET.Element("item")
        ET.SubElement(
            item, "title").text = title
        ET.SubElement(item, "description").text = FeedCreater._escape(
            description)
        # ET.SubElement(item, "summary").text = _escape(i["description"])
        # ET.SubElement(item, "image").text = i['thumbnail']
        ET.SubElement(item, "link").text = webpageUrl
        ET.SubElement(item, "guid").text = webpageUrl
        # ET.SubElement(item, "author").text = i['a3b_info']['user']
        ET.SubElement(item, "pubDate").text = datetime.now().strftime(
            "%a, %d %b %Y %H:%M:%S") + " +0300"
        # <pubDate>Tue, 02 Oct 2016 19:45:02</pubDate>

        url2media = urljoin(
            self.url2content,
            self.feedName)
        url2media = urljoin(
            url2media, self.feedName + '/')
        url2media = urljoin(url2media, self.filename + '.' + fileExtention)

        media_type = 'audio/' + fileExtention
        # content = ET.SubElement(item, "content")
        # content.set("url", url2media)
        # content.set("fileSize", str(i["filesize"]))
        # content.set("type", media_type)

        enclosure = ET.SubElement(item, "enclosure")
        enclosure.set("url", url2media)

        if fileSize != "":
            enclosure.set("length", fileSize)
        enclosure.set("type", media_type)

        channel.append(item)

        ET.indent(tree, space="\t", level=0)
        tree.write(self.path2rss, xml_declaration=True, encoding="utf-8")

    def _escape(text: str):
        text = text.replace("&", "&amp;")
        text = text.replace("<", "&lt;")
        text = text.replace(">", "&gt;")
        text = text.replace("\"", "&quot;")
        return text
