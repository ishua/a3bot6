# scripts/logger.py
"""
Настройка логирования для проекта
"""
import logging

def init_logger(log_level: str) -> logging.Logger:
    """
    Инициализирует логгер с указанным уровнем логирования
    
    Args:
        log_level: Уровень логирования ('DEBUG', 'INFO', 'ERROR', etc)
        
    Returns:
        Настроенный logger объект
    """
    logger = logging.getLogger('finance')
    logger.setLevel(logging.DEBUG)
    
    formatter = logging.Formatter('%(asctime)s | %(levelname)s | %(message)s')
    
    console_handler = logging.StreamHandler()
    console_handler.setLevel(getattr(logging, log_level))
    console_handler.setFormatter(formatter)
    
    logger.addHandler(console_handler)
    
    return logger


logger = init_logger('INFO')
