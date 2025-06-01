import { useEffect, useState } from "react";
import { searchQuery } from "./api";

interface Result {
  doc_id: string;
  title: string;
  url: string;
  snippet: string;
  score: number;
}

export default function SearchPage() {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<Result[]>([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const limit = 10;

  const handleSearch = async () => {
    const data = await searchQuery(query, offset, limit);
    setResults(data.results);
    setTotal(data.totalResults);
  };

  useEffect(() => {
    if (query) handleSearch();
  }, [offset]);

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <h1 className="text-2xl font-bold mb-4">Open Search</h1>

      <div className="flex mb-6">
        <input
          className="border border-gray-300 rounded-l px-4 py-2 w-full"
          placeholder="Search..."
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && handleSearch()}
        />
        <button
          className="bg-blue-600 text-white px-4 py-2 rounded-r"
          onClick={handleSearch}
        >
          Search
        </button>
      </div>

      <p className="mb-2 text-gray-500">{total} results</p>

      <div className="space-y-6">
        {results.map((r) => (
          <div key={r.doc_id} className="border-b pb-4">
            <a href={r.url} className="text-lg text-blue-600 font-medium hover:underline" target="_blank">
              {r.title || r.url}
            </a>
            <p className="text-sm text-gray-500">{r.url}</p>
            <p className="text-gray-800 mt-1">{r.snippet}</p>
            <p className="text-xs text-gray-400 mt-1">Score: {r.score.toFixed(4)}</p>
          </div>
        ))}
      </div>

      {/* Pagination */}
      <div className="flex justify-between mt-6">
        <button
          onClick={() => setOffset(Math.max(0, offset - limit))}
          disabled={offset === 0}
          className="bg-gray-200 px-4 py-2 rounded disabled:opacity-50"
        >
          Previous
        </button>
        <button
          onClick={() => setOffset(offset + limit)}
          disabled={offset + limit >= total}
          className="bg-gray-200 px-4 py-2 rounded disabled:opacity-50"
        >
          Next
        </button>
      </div>
    </div>
  );
}