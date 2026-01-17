# scripts/repo/assets_repo.py
"""
Репозиторий для работы с metadata/assets_metadata.json
Чтение метаданных инструментов
"""
import json
import os
from typing import Dict, List
from logger import logger


def get_asset_metadata(ticker: str, data_dir: str) -> Dict | None:
    """
    Получить метаданные по тикеру
    
    Args:
        ticker: Тикер инструмента
        data_dir: Путь к директории с данными
        
    Returns:
        Словарь с метаданными или None если не найден
    """
    all_assets = get_all_assets(data_dir)
    
    if not all_assets:
        return None
    
    if ticker not in all_assets:
        logger.error(f"ticker {ticker} not found in assets_metadata")
        return None
    
    return all_assets[ticker]


def get_all_assets(data_dir: str) -> Dict:
    """
    Получить все метаданные инструментов
    
    Args:
        data_dir: Путь к директории с данными
    
    Returns:
        Словарь {ticker: metadata} или пустой dict при ошибке
    """
    metadata_file = os.path.join(data_dir, 'metadata', 'assets_metadata.json')
    
    if not os.path.exists(metadata_file):
        logger.error(f"assets_metadata.json not found at {metadata_file}")
        return {}
    
    try:
        with open(metadata_file, 'r', encoding='utf-8') as f:
            data = json.load(f)
        
        return data
        
    except json.JSONDecodeError as e:
        logger.error(f"invalid JSON in assets_metadata.json: {e}")
        return {}
    
    except Exception as e:
        logger.error(f"failed to read assets_metadata.json: {e}")
        return {}


def get_tickers(data_dir: str) -> List[str]:
    """
    Получить список всех тикеров
    
    Args:
        data_dir: Путь к директории с данными
    
    Returns:
        Список тикеров или пустой список при ошибке
    """
    all_assets = get_all_assets(data_dir)
    
    if not all_assets:
        return []
    
    return list(all_assets.keys())
