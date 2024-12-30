import time
from functools import wraps
from typing import Callable, Optional, Dict
import requests

class APIManager:
    def __init__(self):
        self.rate_limits = {}  # Stores rate limit info per service
        self.api_keys = {}     # Stores API keys per service

    def set_api_key(self, service: str, api_key: str):
        self.api_keys[service] = api_key

    def get_api_key(self, service: str) -> Optional[str]:
        return self.api_keys.get(service)

    def rate_limited(self, calls_per_second: float = 1.0):
        def decorator(func: Callable):
            last_called = 0.0

            @wraps(func)
            def wrapper(*args, **kwargs):
                nonlocal last_called
                elapsed = time.time() - last_called
                wait_time = max(0, (1.0 / calls_per_second) - elapsed)
                if wait_time > 0:
                    time.sleep(wait_time)
                last_called = time.time()
                return func(*args, **kwargs)
            return wrapper
        return decorator

    def make_request(self, service: str, url: str, params: Optional[Dict] = None, headers: Optional[Dict] = None):
        # Update headers with API key if available
        api_key = self.get_api_key(service)
        if api_key:
            headers = headers or {}
            headers['Authorization'] = f'Bearer {api_key}'

        # Make the request
        response = requests.get(url, params=params, headers=headers)

        # Update rate limit tracking
        self._update_rate_limits(service, response)

        # Handle errors
        if not response.ok:
            raise APIError(f"API request failed: {response.status_code} - {response.text}")

        return response.json()

    def _update_rate_limits(self, service: str, response: requests.Response):
        # Extract rate limit info from headers
        limit = response.headers.get('X-RateLimit-Limit')
        remaining = response.headers.get('X-RateLimit-Remaining')
        reset = response.headers.get('X-RateLimit-Reset')

        if all([limit, remaining, reset]):
            self.rate_limits[service] = {
                'limit': int(limit),
                'remaining': int(remaining),
                'reset': int(reset)
            }

    def get_rate_limit_info(self, service: str) -> Optional[Dict]:
        return self.rate_limits.get(service)

class APIError(Exception):
    pass