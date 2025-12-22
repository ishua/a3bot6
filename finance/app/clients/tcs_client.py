"""
Клиент для работы с Tinkoff Invest API
"""
from tinkoff.invest import Client, InstrumentIdType, InstrumentStatus
from typing import Optional, Dict
from datetime import datetime


class TCSClient:
    """
    Класс для работы с Tinkoff Invest API
    Поддерживает context manager для автоматического управления соединением
    """
    
    def __init__(self, token):
        """
        Инициализация клиента
        
        Args:
            token (str): API токен Tinkoff Invest
        """
        self.token = token
        self.client = None
        self._client_context = None
    
    def __enter__(self):
        """
        Вход в context manager
        Открывает соединение с API
        """
        self._client_context = Client(self.token)
        self.client = self._client_context.__enter__()
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        """
        Выход из context manager
        Закрывает соединение с API
        """
        if self._client_context:
            self._client_context.__exit__(exc_type, exc_val, exc_tb)
    
    def get_accounts(self):
        """
        Получить список всех счетов пользователя
        
        Returns:
            list[dict]: Список счетов в формате:
                [
                    {
                        'id': '2000320487',
                        'name': 'Брокерский счёт',
                        'type': 1,
                        'status': 2,
                        'opened_date': datetime,
                        'closed_date': datetime
                    },
                    ...
                ]
        """
        response = self.client.users.get_accounts()
        
        if not response.accounts:
            return []
        
        accounts = []
        for account in response.accounts:
            account_dict = {
                'id': account.id,
                'name': account.name,
                'type': account.type,
                'status': account.status,
            }
            
            # Добавляем дополнительные поля если есть
            if hasattr(account, 'opened_date') and account.opened_date:
                account_dict['opened_date'] = account.opened_date
            
            if hasattr(account, 'closed_date') and account.closed_date:
                account_dict['closed_date'] = account.closed_date
            
            accounts.append(account_dict)
        
        return accounts
    
    def get_portfolio(self, account_id):
        """
        Получить портфель (позиции) по счету
        
        Args:
            account_id (str): ID счета
            
        Returns:
            list[dict]: Список позиций в формате:
                [
                    {
                        'figi': 'BBG004730N88',
                        'quantity': 10.0,
                        'current_price': 271.35,
                        'currency': 'usd',
                        'total_value': 2713.5,
                        'instrument_type': 'share'
                    },
                    ...
                ]
        """
        response = self.client.operations.get_portfolio(account_id=account_id)
        
        if not response.positions:
            return []
        
        positions = []
        for position in response.positions:
            # Конвертируем Quotation в float
            quantity = float(position.quantity.units) + float(position.quantity.nano) / 1_000_000_000
            current_price = float(position.current_price.units) + float(position.current_price.nano) / 1_000_000_000
            
            # Пропускаем нулевые позиции
            if quantity == 0:
                continue
            
            position_dict = {
                'figi': position.figi,
                'quantity': quantity,
                'current_price': current_price,
                'currency': position.current_price.currency.lower(),
                'total_value': quantity * current_price,
                'instrument_type': position.instrument_type
            }
            
            positions.append(position_dict)
        
        return positions
    
    def get_instrument_by_figi(self, figi):
        """
        Получить информацию об инструменте по FIGI
        
        Args:
            figi (str): FIGI инструмента
            
        Returns:
            dict: Информация об инструменте:
                {
                    'figi': 'BBG004730N88',
                    'ticker': 'AAPL',
                    'name': 'Apple Inc.',
                    'instrument_type': 'share',
                    'asset_class': 'Stock',
                    'sector': 'Technology',
                    'currency': 'usd'
                }
            или None если инструмент не найден
        """
        
        # Запрос инструмента по FIGI
        response = self.client.instruments.get_instrument_by(
            id_type=InstrumentIdType.INSTRUMENT_ID_TYPE_FIGI,
            id=figi
        )
        
        if not response.instrument:
            return None
        
        instrument = response.instrument
        
        # Маппинг типов инструментов
        instrument_type_mapping = {
            'share': 'Stock',
            'etf': 'ETF',
            'bond': 'Bond',
            'currency': 'Currency',
            'futures': 'Futures',
            'option': 'Option',
        }
        
        # Определяем тип инструмента
        raw_type = instrument.instrument_type.lower() if hasattr(instrument, 'instrument_type') else 'unknown'
        asset_class = instrument_type_mapping.get(raw_type, 'Unknown')
        
        # Извлекаем сектор если есть
        sector = 'Unknown'
        if hasattr(instrument, 'sector') and instrument.sector:
            sector = instrument.sector
        
        # Формируем результат
        result = {
            'figi': figi,
            'ticker': instrument.ticker if hasattr(instrument, 'ticker') else '',
            'name': instrument.name if hasattr(instrument, 'name') else '',
            'instrument_type': raw_type,
            'asset_class': asset_class,
            'sector': sector,
            'currency': instrument.currency.lower() if hasattr(instrument, 'currency') else 'unknown'
        }
        
        return result
            
    
    def get_instrument_by_ticker(self, ticker: str, class_code: str = 'TQBR') -> Optional[Dict]:
        """
        Получить информацию об инструменте по тикеру
        
        Args:
            ticker (str): Тикер инструмента (например, 'SBER', 'FXGD')
            class_code (str): Класс инструмента (по умолчанию 'TQBR' для акций Мосбиржи)
                             'TQBR' - акции
                             'TQTF' - ETF
                             'TQOB' - облигации
        
        Returns:
            dict: Информация об инструменте (структура как в get_instrument_by_figi)
            или None если инструмент не найден
        """

        # Пробуем найти инструмент по тикеру
        # Сначала ищем среди акций (TQBR)
        response = self.client.instruments.find_instrument(query=ticker)
        
        if not response.instruments:
            return None
        
        # Ищем точное совпадение по тикеру
        instrument = None
        for item in response.instruments:
            if item.ticker.upper() == ticker.upper():
                instrument = item
                break
        
        if not instrument:
            # Если точного совпадения нет, берем первый результат
            instrument = response.instruments[0]
        
        # Маппинг типов инструментов
        instrument_type_mapping = {
            'share': 'Stock',
            'etf': 'ETF',
            'bond': 'Bond',
            'currency': 'Currency',
            'futures': 'Futures',
            'option': 'Option',
        }
        
        # Определяем тип инструмента
        raw_type = instrument.instrument_type.lower() if hasattr(instrument, 'instrument_type') else 'unknown'
        asset_class = instrument_type_mapping.get(raw_type, 'Unknown')
        
        # Извлекаем сектор если есть
        sector = 'Unknown'
        if hasattr(instrument, 'sector') and instrument.sector:
            sector = instrument.sector
        
        # Формируем результат
        result = {
            'figi': instrument.figi,
            'ticker': instrument.ticker,
            'name': instrument.name if hasattr(instrument, 'name') else '',
            'instrument_type': raw_type,
            'asset_class': asset_class,
            'sector': sector,
            'currency': instrument.currency.lower() if hasattr(instrument, 'currency') else 'rub'
        }
        
        return result
            
    
    def get_last_price(self, figi: str) -> Optional[float]:
        """
        Получить последнюю цену инструмента
        
        Args:
            figi (str): FIGI инструмента
            
        Returns:
            float: Последняя цена или None если не удалось получить
        """

        response = self.client.market_data.get_last_prices(figi=[figi])
        
        if not response.last_prices:
            return None
        
        last_price = response.last_prices[0].price
        price = float(last_price.units) + float(last_price.nano) / 1_000_000_000
        
        return price
            

    
    def get_usd_rub_rate(self) -> float:
        """
        Получить текущий курс USD/RUB
        
        Returns:
            float: Курс USD/RUB
            
        Raises:
            Exception: Если не удалось получить курс
        """
        # FIGI для USD000UTSTOM (доллар США)
        USD_FIGI = 'BBG0013HGFT4'
        
        response = self.client.market_data.get_last_prices(figi=[USD_FIGI])
        
        if not response.last_prices:
            raise Exception("API не вернул данные по курсу USD/RUB")
        
        last_price = response.last_prices[0].price
        rate = float(last_price.units) + float(last_price.nano) / 1_000_000_000
        
        if rate <= 0:
            raise Exception(f"Получен некорректный курс USD/RUB: {rate}")
        
        return rate
    
    def get_operations(self, account_id: str, from_date: datetime, to_date: Optional[datetime] = None):
        """
        Получить список операций по счету за период
        """
        from tinkoff.invest import OperationType
        
        if to_date is None:
            to_date = datetime.now()
        
        # Запрос операций
        response = self.client.operations.get_operations(
            account_id=account_id,
            from_=from_date,
            to=to_date
        )
        
        if not response.operations:
            return []
        
        operations = []
        
        for operation in response.operations:
            
            # Тип операции (число)
            op_type_value = operation.operation_type if hasattr(operation, 'operation_type') else 0
            
            # Название типа операции
            op_type_name = 'UNKNOWN'
            if op_type_value == OperationType.OPERATION_TYPE_BUY:
                op_type_name = 'BUY'
            elif op_type_value == OperationType.OPERATION_TYPE_SELL:
                op_type_name = 'SELL'
            elif op_type_value == OperationType.OPERATION_TYPE_DIVIDEND:
                op_type_name = 'DIVIDEND'
            elif op_type_value == OperationType.OPERATION_TYPE_COUPON:
                op_type_name = 'COUPON'
            elif op_type_value == OperationType.OPERATION_TYPE_INPUT:
                op_type_name = 'INPUT'
            elif op_type_value == OperationType.OPERATION_TYPE_OUTPUT:
                op_type_name = 'OUTPUT'
            
            # Статус (1 = Executed)
            state_value = operation.state if hasattr(operation, 'state') else 0
            state_name = 'EXECUTED' if state_value == 1 else 'OTHER'
            
            # FIGI
            figi = operation.figi if hasattr(operation, 'figi') and operation.figi else ''
            
            # Дата
            date = operation.date if hasattr(operation, 'date') else None
            
            # Количество
            quantity = operation.quantity if hasattr(operation, 'quantity') else 0
            
            # Payment (сумма операции)
            payment = 0
            if hasattr(operation, 'payment') and operation.payment:
                payment = float(operation.payment.units) + float(operation.payment.nano) / 1_000_000_000
            
            # Цена
            price = 0
            if hasattr(operation, 'price') and operation.price:
                price = float(operation.price.units) + float(operation.price.nano) / 1_000_000_000
            
            # Комиссия
            commission = 0
            if hasattr(operation, 'commission') and operation.commission:
                commission = float(operation.commission.units) + float(operation.commission.nano) / 1_000_000_000
                commission = abs(commission)
            
            # Валюта
            currency = 'unknown'
            if hasattr(operation, 'payment') and operation.payment and hasattr(operation.payment, 'currency'):
                currency = operation.payment.currency.lower()
            
            # ID операции
            op_id = operation.id if hasattr(operation, 'id') else ''
            
            operation_dict = {
                'id': op_id,
                'date': date,
                'type': op_type_name,
                'state': state_name,
                'figi': figi,
                'quantity': quantity,
                'price': price,
                'payment': payment,
                'currency': currency,
                'commission': commission,
            }
            
            operations.append(operation_dict)
        
        return operations
    

    def get_operations_raw(self, account_id: str, from_date: datetime, to_date: Optional[datetime] = None):
        """
        Получить список операций по счету за период (сырые данные)
        Сохраняет максимум информации без преобразований
        """
        if to_date is None:
            to_date = datetime.now()
        
        # Запрос операций
        response = self.client.operations.get_operations(
            account_id=account_id,
            from_=from_date,
            to=to_date
        )
        
        if not response.operations:
            return []
        
        operations = []
        
        for operation in response.operations:
            # Сохраняем все поля как есть
            op_dict = {
                'id': operation.id if hasattr(operation, 'id') else '',
                'parent_operation_id': operation.parent_operation_id if hasattr(operation, 'parent_operation_id') else '',
                'currency': operation.currency if hasattr(operation, 'currency') else '',
                'payment': self._quotation_to_float(operation.payment) if hasattr(operation, 'payment') else 0,
                'price': self._quotation_to_float(operation.price) if hasattr(operation, 'price') else 0,
                'state': operation.state if hasattr(operation, 'state') else 0,
                'quantity': operation.quantity if hasattr(operation, 'quantity') else 0,
                'quantity_rest': operation.quantity_rest if hasattr(operation, 'quantity_rest') else 0,
                'figi': operation.figi if hasattr(operation, 'figi') else '',
                'instrument_type': operation.instrument_type if hasattr(operation, 'instrument_type') else '',
                'date': operation.date.isoformat() if hasattr(operation, 'date') and operation.date else None,
                'type': operation.type if hasattr(operation, 'type') else '',
                'operation_type': operation.operation_type if hasattr(operation, 'operation_type') else 0,
                'trades': [],
            }
            
            # Комиссия
            if hasattr(operation, 'commission') and operation.commission:
                op_dict['commission'] = self._quotation_to_float(operation.commission)
            else:
                op_dict['commission'] = 0
            
            # Доходность
            if hasattr(operation, 'yield') and operation.yield_:
                op_dict['yield'] = self._quotation_to_float(operation.yield_)
            else:
                op_dict['yield'] = 0
            
            # Accumulated coupon interest
            if hasattr(operation, 'accrued_int') and operation.accrued_int:
                op_dict['accrued_int'] = self._quotation_to_float(operation.accrued_int)
            else:
                op_dict['accrued_int'] = 0
            
            # Trades (детали сделок)
            if hasattr(operation, 'trades') and operation.trades:
                for trade in operation.trades:
                    trade_dict = {
                        'trade_id': trade.trade_id if hasattr(trade, 'trade_id') else '',
                        'date': trade.date_time.isoformat() if hasattr(trade, 'date_time') and trade.date_time else None,
                        'quantity': trade.quantity if hasattr(trade, 'quantity') else 0,
                        'price': self._quotation_to_float(trade.price) if hasattr(trade, 'price') else 0,
                    }
                    op_dict['trades'].append(trade_dict)
            
            operations.append(op_dict)
        
        return operations


    def _quotation_to_float(self, quotation):
        """
        Конвертировать Quotation в float
        """
        if not quotation:
            return 0.0
        return float(quotation.units) + float(quotation.nano) / 1_000_000_000
