from models.search_result import SearchResult
from typing import List
from utils.api_manager import APIManager
from datetime import datetime
import trafilatura
import os

class TrafilaturaAdapter:
    def __init__(self, api_manager: APIManager):
        self.api_manager = api_manager
        self.use_local = os.getenv('TRAFILATURA_API_KEY') is None
        self.base_url = "https://api.trafilatura.com/v1" if not self.use_local else None

    def search(self, query: str) -> List[SearchResult]:
        try:
            if self.use_local:
                return self._search_local(query)
            else:
                return self._search_api(query)
        except Exception as e:
            print(f"Error querying trafilatura: {str(e)}")
            return []

    def _search_api(self, query: str) -> List[SearchResult]:
        # Make API request
        endpoint = f"{self.base_url}/search"
        params = {
            'query': query,
            'limit': 10,
            'lang': 'en'
        }
        data = self.api_manager.make_request(
            service='trafilatura',
            url=endpoint,
            params=params
        )

        # Transform results
        results = []
        for item in data.get('results', []):
            # Extract main content using trafilatura
            downloaded = trafilatura.fetch_url(item['url'])
            content = trafilatura.extract(downloaded) or item.get('description', '')

            results.append(SearchResult(
                title=item.get('title', 'News Article'),
                description=content[:500],  # Limit description length
                source="trafilatura",
                url=item.get('url'),
                relevance=float(item.get('relevance_score', 0.0)),
                metadata={
                    'source_domain': item.get('domain'),
                    'language': item.get('language', 'en'),
                    'date': item.get('date', datetime.now().isoformat()),
                    'word_count': len(content.split())
                }
            ))
        return results

    def _search_local(self, query: str) -> List[SearchResult]:
        # Implement local search logic here
        # This could involve searching through local documents or databases
        return [
            SearchResult(
                title="Local News Article",
                description="Example result from local trafilatura search",
                source="trafilatura",
                relevance=0.6
            )
        ]