from dataclasses import dataclass
from typing import Optional

@dataclass
class SearchResult:
    title: str
    description: str
    source: str
    url: Optional[str] = None
    relevance: float = 0.0
    metadata: Optional[dict] = None