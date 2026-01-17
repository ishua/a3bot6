"""
Настройка логирования для проекта
"""
import logging

# Создаём логгер сразу, но без handler'ов
logger = logging.getLogger('finance')

def init_logger(log_level: str = 'INFO') -> None:
    """Инициализировать логгер. Вызывать один раз в main."""
    if logger.handlers:
        return
    
    logger.setLevel(logging.DEBUG)
    
    formatter = logging.Formatter('%(asctime)s | %(levelname)s | %(message)s')
    
    console_handler = logging.StreamHandler()
    console_handler.setLevel(getattr(logging, log_level))
    console_handler.setFormatter(formatter)
    
    logger.addHandler(console_handler)
