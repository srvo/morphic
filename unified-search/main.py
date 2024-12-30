from adapters.physrisk import PhysRiskAdapter
from adapters.trafilatura import TrafilaturaAdapter
from adapters.watchman import WatchmanAdapter
from models.search_result import SearchResult
from utils.api_manager import APIManager
from typing import List
import os

class UnifiedSearch:
    def __init__(self):
        # Initialize API manager
        self.api_manager = APIManager()

        # Initialize adapters with local configurations
        self.physrisk = PhysRiskAdapter(self.api_manager)
        self.trafilatura = TrafilaturaAdapter(self.api_manager)
        self.watchman = WatchmanAdapter(self.api_manager)

    def search(self, query: str) -> List[SearchResult]:
        import logging
        logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
        results = []
        try:
            results.extend(self.physrisk.search(query))
        except Exception as e:
            logging.error(f"Error querying physrisk API: {e}")
        try:
            results.extend(self.trafilatura.search(query))
        except Exception as e:
            logging.error(f"Error querying trafilatura API: {e}")
        try:
            results.extend(self.watchman.search(query))
        except Exception as e:
            logging.error(f"Error querying watchman API: {e}")
        return self._rank_results(results)

    def _rank_results(self, results: List[SearchResult]) -> List[SearchResult]:
        # Enhanced ranking logic with None handling
        return sorted(results, key=lambda x: (
            x.relevance,
            -len(x.metadata.get('entities', [])) if x.source == 'watchman' and x.metadata else 0,
            x.metadata.get('word_count', 0) if x.source == 'trafilatura' and x.metadata else 0
        ), reverse=True)

    def get_rate_limit_info(self) -> dict:
        return {
            'physrisk': self.api_manager.get_rate_limit_info('physrisk'),
            'trafilatura': self.api_manager.get_rate_limit_info('trafilatura'),
            'watchman': self.api_manager.get_rate_limit_info('watchman')
        }

if __name__ == "__main__":
    search_tool = UnifiedSearch()
    query = "test"
    results = search_tool.search(query)
    
    # Print results
    for result in results:
        print(f"\n[{result.source}] {result.title}")
        print(f"  {result.description[:200]}...")
        print(f"  URL: {result.url}")
        print(f"  Relevance: {result.relevance:.2f}")
    
    # Print rate limit info
    print("\nRate Limit Information:")
    for service, info in search_tool.get_rate_limit_info().items():
        if info:
            print(f"{service}: {info['remaining']}/{info['limit']} (resets in {info['reset']}s)")
