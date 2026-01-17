# scripts/repo/sync_repo.py
"""
Репозиторий для работы с metadata/last_sync.json
Отслеживание состояния синхронизации данных
"""
import json
import os
from datetime import datetime
from typing import List, Dict
from logger import logger


def update_positions_sync(positions: List[Dict], data_dir: str, source: str = 'manual') -> bool:
    """
    Обновляет секцию positions в last_sync.json
    
    Args:
        positions: Список позиций
        data_dir: Путь к директории с данными
        source: Источник данных (tcs-api, manual)
        
    Returns:
        True если успешно, False при ошибке
    """
    try:
        sync_file = os.path.join(data_dir, 'metadata', 'last_sync.json')
        sync_data = _load_sync_data(sync_file)
        timestamp = datetime.now().isoformat()
        
        accounts_data = _build_accounts_data(positions, source, timestamp)
        
        sync_data['positions'] = {
            'last_updated': timestamp,
            'status': 'success',
            'total_records': len(positions),
            'accounts': accounts_data
        }
        
        _save_sync_data(sync_file, sync_data)
        logger.debug(f"updated last_sync.json: {len(positions)} positions")
        return True
        
    except Exception as e:
        logger.error(f"failed to update last_sync.json: {e}")
        return False


def _load_sync_data(sync_file: str) -> Dict:
    """
    Загружает текущий last_sync.json или создает пустую структуру
    
    Args:
        sync_file: Путь к файлу last_sync.json
        
    Returns:
        Dict с данными синхронизации
    """
    if os.path.exists(sync_file):
        try:
            with open(sync_file, 'r', encoding='utf-8') as f:
                return json.load(f)
        except:
            pass
    
    return {
        'positions': {},
        'transactions': {},
        'positions_history': {}
    }


def _save_sync_data(sync_file: str, data: Dict) -> None:
    """
    Сохраняет данные в last_sync.json
    
    Args:
        sync_file: Путь к файлу last_sync.json
        data: Данные для сохранения
    """
    os.makedirs(os.path.dirname(sync_file), exist_ok=True)
    
    with open(sync_file, 'w', encoding='utf-8') as f:
        json.dump(data, f, indent=2, ensure_ascii=False)


def _build_accounts_data(positions: List[Dict], source: str, timestamp: str) -> Dict:
    """
    Группирует позиции по счетам и собирает статистику
    
    Args:
        positions: Список позиций
        source: Источник данных
        timestamp: Временная метка
        
    Returns:
        Dict с данными по счетам
    """
    accounts = {}
    
    for pos in positions:
        account = pos['account']
        
        if account not in accounts:
            accounts[account] = {
                'broker': _detect_broker(account),
                'last_updated': timestamp,
                'records_count': 0,
                'value_rub': 0,
                'status': 'success',
                'source': source
            }
        
        accounts[account]['records_count'] += 1
        accounts[account]['value_rub'] += pos['value_in_rub']
    
    return accounts


def _detect_broker(account: str) -> str:
    """
    Определяет брокера по системному названию счета
    
    Args:
        account: Системное название счета
        
    Returns:
        str: Код брокера (TCS, SBER, FINEXP, UNKNOWN)
    """
    if account in ['GENERAL', 'FAMILY', 'IIS_2025', 'IIS_122025', 'InvestBox']:
        return 'TCS'
    elif account == 'SBER':
        return 'SBER'
    elif account == 'FINEXP':
        return 'FINEXP'
    else:
        logger.error(f"unknown account for broker detection: '{account}'")
        return 'UNKNOWN'
