from transmission_rpc import Client

class Tr:
    def __init__( self, host: str, port: int, download_dir: str):
        self.client = Client(host=host, port=port)
        self.download_dir = download_dir

    def add_torrent(self, torrent_url: str, folder_path: str) -> str:
        _download_dir = self.download_dir
        _download_dir += folder_path
        try:
            t = self.client.add_torrent(torrent=torrent_url, download_dir=_download_dir)
        except Exception as e:
            return e.__str__()

        return str(t.id) + " " + t.name

    def list_torrents(self) -> str:

        try:
            tlist = self.client.get_torrents()
        except Exception as e:
            return e.__str__()
        if len(tlist) == 0:
            return "no torrents"
        ret = "torrent list \n"
        for t in tlist:
            ret += str(t.id)
            ret +=  "-" + t.name
            ret += "-" + t.status
            ret += "-" + str(t.progress)
            ret += "-" + '{:.3f}'.format(t.rate_download/1024/1024)
            ret += "\n"
        return ret

    def del_torrent(self, torrent_id: int) -> str:
        try:
            t = self.client.remove_torrent(torrent_id)
        except Exception as e:
            return e.__str__()
        return "ok"