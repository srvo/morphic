import pytest
from unittest.mock import patch
from unified_search.main import UnifiedSearch
from unified_search.models.search_result import SearchResult

@pytest.fixture
def search_tool():
    return UnifiedSearch()

def test_search_integration(search_tool):
    with patch('unified_search.adapters.physrisk.PhysRiskAdapter.search') as mock_physrisk, \
         patch('unified_search.adapters.watchman.WatchmanAdapter.search') as mock_watchman, \
         patch('unified_search.adapters.trafilatura.TrafilaturaAdapter.search') as mock_trafilatura:
        
        # Mock adapter responses
        mock_physrisk.return_value = [
            SearchResult(
                title="Test PhysRisk",
                description="Test description",
                source="physrisk",
                relevance=0.8
            )
        ]
        
        mock_watchman.return_value = [
            SearchResult(
                title="Test Watchman",
                description="Test description",
                source="watchman",
                relevance=0.9
            )
        ]
        
        mock_trafilatura.return_value = [
            SearchResult(
                title="Test Trafilatura",
                description="Test description",
                source="trafilatura",
                relevance=0.7
            )
        ]
        
        # Execute search
        results = search_tool.search("test query")
        
        # Verify results
        assert len(results) == 3
        assert results[0].source == "watchman"  # Highest relevance
        assert results[1].source == "physrisk"
        assert results[2].source == "trafilatura"

def test_search_with_empty_results(search_tool):
    with patch('unified_search.adapters.physrisk.PhysRiskAdapter.search') as mock_physrisk, \
         patch('unified_search.adapters.watchman.WatchmanAdapter.search') as mock_watchman, \
         patch('unified_search.adapters.trafilatura.TrafilaturaAdapter.search') as mock_trafilatura:
        
        # Mock empty responses
        mock_physrisk.return_value = []
        mock_watchman.return_value = []
        mock_trafilatura.return_value = []
        
        results = search_tool.search("test query")
        assert len(results) == 0