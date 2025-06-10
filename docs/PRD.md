## Project Summary: OpenSearch


OpenSearch is a full-stack, distributed web crawler and search engine system designed to demonstrate the capabilities of large-scale, concurrent web crawling, content indexing, and real-time keyword-based information retrieval. It is implemented using modern technologies with an emphasis on concurrency, modularity, and production-readiness.


**This project is built using:**

• Golang for the crawler, backend API, and indexing engine.

• TypeScript + React + TailwindCSS + Axios for the frontend UI.

• MongoDB for content and index storage.

• Redis for distributed queueing and visited URL tracking.

• Docker for containerized deployment of all services.

• GitHub Actions for CI/CD automation.

• Prometheus + Grafana for real-time observability and monitoring. \*In Progress\*

• (Optional) Kafka, Kubernetes, and a vector DB for scale and semantic search capabilities.


---


### Product Requirements Document (PRD)


1. **Goals**

• Build a concurrent, scalable web crawler.

• Create a robust inverted index with TF-IDF ranking.

• Develop a production-grade search engine frontend and backend.

• Ensure the system is ethical, observable, fault-tolerant, and extensible.

• Showcase advanced features like language detection, semantic enhancements, and DevOps integrations.

---


2. **Core Functional Requirements**


2.1. Web Crawler

• Implemented in Go using colly and goquery.

• Supports:

- Breadth-first search (BFS) crawling logic.

- Concurrency using goroutines and worker pools.

- Thread-safe queue with a mutex.

- URL normalization (e.g., canonicalization, lowercase, query stripping).

- Deduplication via visited set and content hashing (e.g., SHA-1).

• Politeness:

- Rate limiting via colly.LimitRule

- Respect robots.txt via colly.WithRobotsTxt()

• Filter:

- Non-English content using whatlanggo.

- Pages with content size exceeding a configurable max byte threshold.

• Stores:

- Raw HTML, page title, URL, cleaned text content.


2.2. Persistent Queue + Crash Recovery

- Redis or MongoDB-backed queue implementation.

- Stores visited URLs and queued URLs for recovery after failure.

- Snapshot progress periodically or after every N pages.


---


3. **Indexing System**


3.1. Inverted Index

• Tokenizes cleaned content and builds inverted index {word -> [docIDs]}

• Tracks:

- Term frequency per document.

- Document frequency per term.

• Indexed metadata:

- Page URL

- Title

- Content snippet(s)

- Word positions


3.2. Ranking

• Use TF-IDF scoring:

- TF = #term occurrences / total terms

- IDF = log(total documents / documents containing term)

• Support boosting based on:

- Term in title

- URL relevance


3.3. Snippets

• Generate text snippets where query terms appear (sentence/paragraph level).

• Store during indexing to avoid recomputation.


---


4. Search Engine API (Go)

• RESTful API using Go’s standard library or a framework (e.g., Gin or Fiber).

• Endpoints:

- /search?q=query returns ranked results with:

  - Title

  - URL

  - Snippet

  - Score

- /status returns crawl stats and system info.

• Optional: /suggest?q=partial for auto-complete.


⸻


5. **Frontend Interface** (React + TS + TailwindCSS)

• Pages:

- Home/search page with query bar.

- Results page with paginated, ranked results.

- Site stats or “About” page.

• Features:

- Search suggestions (auto-complete)

- Query modifiers (e.g., intitle:, site:)

- Clean, modern UI with TailwindCSS

- Integration with API via Axios.


---


6. **Observability and Monitoring**

• Instrument crawler and API with Prometheus metrics:

- Pages/sec

- Queue length

- Number of English pages vs. skipped

- Errors, failed fetches, etc.

• Expose /metrics endpoint.

- Use Grafana to build:

- Dashboard showing crawl status

- Alerts for high error rates


---


7. **DevOps Infrastructure**


7.1. Docker

• Containerize:

- Crawler

- Backend API

- Frontend

- MongoDB

- Redis

• docker-compose file to run everything locally.


7.2. CI/CD (GitHub Actions)

• Linting and testing on:

- Go backend

- TypeScript frontend

- Build and test Docker images.

- Optionally deploy to cloud (e.g., Render, DigitalOcean, or Fly.io).


7.3. Optional Kubernetes Support

• Helm chart or manifest to deploy all services on K8s.

• Set up Horizontal Pod Autoscaler for crawler and backend.


8. Advanced/Stretch Features



**Content Hashing**:
Prevent duplicate content by hashing cleaned content and comparing

**Search Operators**:
Support intitle:, site:, exact phrase search

**Semantic Search**:
Use embedding models (e.g., OpenAI, SentenceTransformers) with a vector DB (Weaviate/Qdrant)

**Distributed Crawling**:
Use Kafka + Redis to coordinate multiple crawler nodes

**Page Change Detection**:
Recrawl periodically and flag updated pages

**Auto-suggestions**:
Based on previous queries or indexed terms



9. **Non-Functional Requirements**

• **Performance**: Crawl speed > 5 pages/sec with concurrency

• **Scalability**: Can support distributed crawling and sharded indexing

• **Reliability**: Persistent queue to resume progress after crash

• **Security**: Avoid crawling login-protected or sensitive pages

• **Maintainability**: Modular code and proper logging


⸻


11. **Deliverables**

• Fully functional web crawler

• Inverted index-based search engine with ranking

• Web-based UI with keyword-based query

• Dockerized services with orchestration

• GitHub repository with CI/CD workflows

• System architecture documentation

• Prometheus + Grafana dashboard \*In Progress\*

• (Optional) Kubernetes deployment scripts 