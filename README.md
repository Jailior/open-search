#  OpenSearch

**OpenSearch** is distributed web search engine made for educational and hobbyist purposes. OpenSearch is follows the Google search architecture including microservices for crawling, indexing and querying.

OpenSearch is available live now at [opensearch.app](http://3.144.8.239:3000)



## Features

- **Web Search**: OpenSearch provides a simple responsive web UI for querying the web with results returned and ranked by the PageRank and TF-IDF algorithms.
- **Look at these!**: is a simple metrics about page providing live information on the latest number of pages crawled, total searches served, and many more!

## Architecture

**OpenSearch** is built around a microservice architecture where all services are independently dockerized and deployable. This modular design allows for flexible changes and scaling of each service. 

### Architecture Diagram

This architecture diagram provides a good overview of the architecture achieved my OpenSearch:

```
frontend (React/TS)
    │
    ▼
[API Service] ─────► MongoDB (pages, index, PageRank)
    ▲                   ▲
    |          ─────────|
    │          │ 
[PageRank] [Indexer] ◄──── Redis Stream ◄──── [Crawler]
                                                  ▲ 
                                                  |
                                    MongoDB Redis Queue/Set
```


### Services

- **Web Crawler**: A concurrent, English-only crawler with Redis-backed persistence crawls the web in a breath-first-search manner.
- **Indexer**: Inverted index builder using Redis stream message passing and MongoDB for document and index storage, runs in parallel and concurrently with the crawler.
- **PageRank**: Builds a link graph from crawled pages and computes PageRank. Scores are stored in MongoDB and used during query ranking.
- **Query Engine**: Query-time scoring using TF-IDF + PageRank blending for improved relevance and authority.
- **Search API**: Fast and stateless REST API built with Go and Gin.
- **Client**: Responsive React + TypeScript + Tailwind UI with snippet highlighting, pagination, and search history tracking.
- **Metrics**: MongoDB-backed crawler telemetry (pages crawled, skipped, queue size, etc.) with future Prometheus integration.
- **CI/CD**: GitHub Actions pipeline builds and pushes all services to Docker Hub.

## Key Performance Achievements

- Batch fetching postings and raw documents reduced query latency by more than 700%.
- An average of 200 pages crawled per minute with a peak of 400 per/min by parallelizing crawler and indexer among concurrent workers.
- PageRank + TF-IDF normalization and blended score significantly improved results quality.


## Tech Stack

- **Backend**: Go (Gin, Colly, Mongo driver, Redis)

- **Frontend**: React + TypeScript + TailwindCSS

- **Database**: MongoDB

- **Queue/Cache**: Redis (Streams and Lists)

- **Infrastructure**: Docker, Docker Compose, AWS Lightsail

- **CI/CD**: GitHub Actions + Docker Hub deployment



## Repo Structure

```
.
├── backend/
│ ├── cmd/
│ │ ├── api/                # Search backend
│ │ ├── crawler/            # Web crawler service
│ │ ├── indexer/            # Inverted index builder
│ │ └── pagerank/           # PageRank processor
│ └── internal/             # Shared packages
├── frontend/client/        # React + TS frontend
├── docker-compose.yml      # Orchestration
└── .github/workflows/      # GitHub Actions (CI/CD)
```


## Getting Started

#### Running Locally

```bash
# Clone the repo
git clone https://github.com/Jailior/open-search.git
cd open-search

# Build and run services
docker compose up --build

# Start and stop each of the services as needed
docker compose up <service-name>
docker compose stop <service-name>

# Example: Passing flags to crawler
docker compose --profile optional run --rm crawler ./crawler --reset --workers 4 
```

You can also *manually seed* the crawler:
```bash
docker exec -it <redis-container> redis-cli
> LPUSH url_queue "https://www.google.com/"
```

You will also need to set up any environment variables which may differ from the current setup:

- `MONGODB_URI`: `mongodb://localhost:27017`
- `REDIS_ADDR`: `localhost:6379`


## License

MIT License, use freely for educational and personal purposes.