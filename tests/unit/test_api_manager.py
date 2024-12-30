import pytest
from unittest.mock import patch
from unified_search.utils.api_manager import APIManager, APIError

def test_api_manager_initialization(api_manager):
    assert api_manager.rate_limits == {}
    assert api_manager.api_keys == {}

def test_set_and_get_api_key(api_manager):
    api_manager.set_api_key('test_service', 'test_key')
    assert api_manager.get_api_key('test_service') == 'test_key'

def test_make_request_success(api_manager):
    with patch('requests.get') as mock_get:
        mock_get.return_value.status_code = 200
        mock_get.return_value.json.return_value = {'data': 'test'}
        mock_get.return_value.headers = {
            'X-RateLimit-Limit': '100',
            'X-RateLimit-Remaining': '99',
            'X-RateLimit-Reset': '60'
        }
        
        result = api_manager.make_request('test_service', 'http://test.com')
        assert result == {'data': 'test'}
        assert api_manager.rate_limits['test_service'] == {
            'limit': 100,
            'remaining': 99,
            'reset': 60
        }

def test_make_request_failure(api_manager):
    with patch('requests.get') as mock_get:
        mock_get.return_value.status_code = 400
        mock_get.return_value.text = 'Bad Request'
        
        with pytest.raises(APIError):
            api_manager.make_request('test_service', 'http://test.com')

def test_rate_limited_decorator(api_manager):
    @api_manager.rate_limited(calls_per_second=1)
    def test_func():
        return True
        
    assert test_func() is True