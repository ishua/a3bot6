import yaml
import sys
from datetime import datetime

class Conf():
    def __init__(self):
        self.conf = {}
        with open("conf/fin_config.yaml") as f:
            self.conf = yaml.load(f, Loader=yaml.FullLoader)

    @property
    def task_type(self) -> int:
        return self.conf.get("taskType", 6)

    @property
    def mcore_addr(self) -> str:
        return self.conf.get("mcoreAddr", "http://localhost:8080")

    @property
    def mcore_secret(self) -> str:
        return self.conf.get("mcoreSecret", "test")
    
    @property
    def data_path(self) -> str:
        return self.conf.get("dataPath", "data")
    
    @property
    def log_level(self) -> str:
        return self.conf.get("logLevel", "INFO")
    
    @property 
    def account_mapping(self) -> dict:
        return self.conf.get("accountMapping", {})
    
    def history_snapshots_days(self) -> list:
        return self.conf.get("historySnapshitsDays", [])

    def should_save_history_snapshot(self) -> bool:
        """
        Проверяет, нужно ли сохранять снимок в positions_history
        
        Returns:
            bool: True если текущий день в HISTORY_SNAPSHOT_DAYS
        """
        current_day = datetime.now().day
        return current_day in self.history_snapshots_days

    def map_account_name(self, name: str) -> str:
        """
        Преобразует название счета из TCS API в системное название
        
        Args:
            tcs_name: Название счета из TCS API
            
        Returns:
            str: Системное название счета или исходное если маппинг не найден
        """
        mapped_name = self.account_mapping.get(name)
        
        if mapped_name is None:
            from logger import logger
            logger.error(f"unmapped TCS account name: '{name}'")
            return name
        
        return mapped_name