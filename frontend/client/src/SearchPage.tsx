import { useEffect, useState } from "react";
import { LuSearchCheck } from "react-icons/lu";

import styles from "./SearchPage.module.css"
import { fetchMetrics, searchQuery } from "./api";
import ResultCard, { type Result } from "./ResultCard";
import MetricsCard, {type Metrics } from "./MetricsCard";
import { FaAngleLeft } from "react-icons/fa";


export default function SearchPage() {
  const [query, setQuery] = useState("");
  const [hasSearched, setHasSearched] = useState(false);

  const [results, setResults] = useState<Result[]>([]);
  const [metrics, setMetrics] = useState<Metrics>();
  const [showMetrics, setShowMetrics] = useState(false);

  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const limit = 15;

  const handleSearch = async () => {
    const data = await searchQuery(query, offset, limit);
    setResults(data.results);
    setTotal(data.totalResults);
    setHasSearched(true);
    setShowMetrics(false);
  };

  const handleMetrics = async () => {
    const data = await fetchMetrics();
    setMetrics(data.metrics);
    setShowMetrics(true);
  }

  useEffect(() => {
    if (query) handleSearch();
  }, [offset]);

  useEffect(() => {
  if (query === "") {
    setHasSearched(false);
    setResults([]);
    setTotal(0);
    setOffset(0);
  }
}, [query]);

  return (
    <div
    className={`${styles.container} ${
    hasSearched ? "items-start pt-12 ml-20" : "items-center justify-center"
    }
    ${
        (hasSearched && total == 0) ? "min-h-screen" : ""
    } 
    flex flex-col transition-all duration-200`}
    >
    
      {/* Title */}
      <div className="sticky top-0 z-10 bg-transparent">
        <h1 
          className={`${styles.title} ${hasSearched ? styles.titleSmall : styles.titleLarge}`}
          onClick={() => {
              setQuery("");
              setHasSearched(false);
              setOffset(0);
              setResults([]);
              setTotal(0);
              setShowMetrics(false);
          }}
        >
          <LuSearchCheck className="inline text-green-reseda" /> 
          Open<span className="text-green-reseda">Search</span>
          </h1>
        </div>

      {/* Search Bar */}
      {!showMetrics && (
        <div className={`${styles.searchBar} sticky top-16 z-10 bg-transparent`}>
            <input
            className={`${styles.input} w-full transition-all duration-300`}
            placeholder="Search..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleSearch()}
            />

            {hasSearched && (
            <button
            className={`${styles.button} ml-4`}
            onClick={handleSearch}
            >
            Search
            </button>
            )}
        </div>
      )}

      {/* Buttons below input */}
      {!hasSearched && !showMetrics && (
      <>
        <div className="flex space-x-4 mb-8 justify-center">
            <button
            className={styles.button}
            onClick={handleSearch}
            >
            Search
            </button>
            <button
            className={styles.button}
            onClick={handleMetrics}
            >
            Look at these!
            </button>
        </div>
            <div>
                <p className={`${styles.factText} mt-10`}>Search from ~26,000 webpages!</p>
        </div>
      </>
      )}

      
      {/* Total results found and results list */}
        <div
            className={`transition-opacity ${hasSearched && (total != 0) ? "opacity-100" : "opacity-0" }`}
        >
        <p className={`${styles.totalResults} duration-0`}>Showing {total} results for {query}</p>
        
        <div className={`${styles.resultList} duration-500`}>
                {results.map((r) => (
                    <ResultCard key={r.doc_id} result={r} />
                ))}
            </div>
        </div>

    {/* Metrics */}
    {metrics && !hasSearched && showMetrics && (
        <>
        <div className="w-full max-w-4xl mx-auto">
          <button className={`${styles.backButton} mb-0 ml-36`}
            onClick={() => {
              setShowMetrics(false);
            }}>
            <FaAngleLeft></FaAngleLeft>
          </button>
          </div>
          <div
              className={`transition-opacity duration-500 ${showMetrics ? "opacity-100" : "opacity-0 invisible"}`}
          >
              <div className={styles.resultList}>
              <MetricsCard metrics={metrics} />
              </div>
          </div>
        </>
    )}

    {/* No results found block */}
    <div
        className={`transition-opacity duration-500 ${hasSearched && (total == 0)? "opacity-100" : "opacity-0" }`}
    >
        <p className={`${styles.totalResults}`}> No matching results T-T</p>
    </div>

    {/* Pagination */}
    {(hasSearched && (total != 0)) && (
      <div className={styles.pagination}>
        <button
          onClick={() => {
            setOffset(Math.max(0, offset - limit));
            window.scrollTo({ top: 0, behavior: "smooth" });
          }}
          disabled={offset === 0}
          className={styles.pageButton}
        >
          Previous
        </button>
        <button
          onClick={() => {
            setOffset(offset + limit);
            window.scrollTo({ top: 0, behavior: "smooth" });
          }}
          disabled={offset + limit >= total}
          className={`${styles.pageButton} ml-3`}
        >
          Next
        </button>
      </div>
    )}
    </div>
  );
}