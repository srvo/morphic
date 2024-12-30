from models.search_result import SearchResult
from typing import List
from utils.api_manager import APIManager
from datetime import datetime
import os

class PhysRiskAdapter:
    def __init__(self, api_manager: APIManager):
        self.api_manager = api_manager
        self.base_url = os.getenv('PHYSRISK_BASE_URL', 'http://localhost:8000/api/v1')

    def search(self, query: str) -> List[SearchResult]:
        try:
            # Make API request
            endpoint = f"{self.base_url}/risks/search"
            params = {
                'query': query,
                'limit': 10,
                'sort': 'relevance'
            }
            data = self.api_manager.make_request(
                service='physrisk',
                url=endpoint,
                params=params
            )

            # Transform results
            results = []
            for item in data.get('results', []):
                results.append(SearchResult(
                    title=item.get('title', 'Climate Risk Data'),
                    description=item.get('description', ''),
                    source="physrisk",
                    url=item.get('url'),
                    relevance=float(item.get('relevance_score', 0.0)),
                    metadata={
                        'location': item.get('location'),
                        'risk_type': item.get('risk_type'),
                        'date': item.get('date', datetime.now().isoformat())
                    }
                ))
            return results

        except Exception as e:
            # Log error and return empty list
            print(f"Error querying physrisk API: {str(e)}")
            return []