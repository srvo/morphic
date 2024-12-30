from models.search_result import SearchResult
from typing import List
from utils.api_manager import APIManager
from datetime import datetime
import os

class WatchmanAdapter:
    def __init__(self, api_manager: APIManager):
        self.api_manager = api_manager
        self.base_url = os.getenv('WATCHMAN_BASE_URL', 'http://localhost:8080/api/v1')

    def search(self, query: str) -> List[SearchResult]:
        try:
            # Make API request
            endpoint = f"{self.base_url}/sanctions/search"
            params = {
                'query': query,
                'limit': 10,
                'sort': 'relevance'
            }
            data = self.api_manager.make_request(
                service='watchman',
                url=endpoint,
                params=params
            )

            # Transform results
            results = []
            for item in data.get('sanctions', []):
                results.append(SearchResult(
                    title=item.get('name', 'Sanction Record'),
                    description=item.get('reason', ''),
                    source="watchman",
                    url=item.get('source_url'),
                    relevance=float(item.get('relevance_score', 0.0)),
                    metadata={
                        'country': item.get('country'),
                        'type': item.get('type'),
                        'issued_date': item.get('issued_date'),
                        'expiration_date': item.get('expiration_date'),
                        'entities': item.get('entities', [])
                    }
                ))
            return results

        except Exception as e:
            # Log error and return empty list
            print(f"Error querying watchman API: {str(e)}")
            return []