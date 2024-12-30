import pytest
from unified_search.utils.api_manager import APIManager

@pytest.fixture
def api_manager():
    return APIManager()