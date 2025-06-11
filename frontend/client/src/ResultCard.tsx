import React from "react";
import styles from "./SearchPage.module.css";

// Standard interface for a page result
export interface Result {
  doc_id: string;
  title: string;
  url: string;
  snippet: string;
  score: number;
}

interface Props {
    result: Result;
}

// Result card component, requires Result instance
const ResultCard: React.FC<Props> = ({ result }) => {
    return (
    <div className={styles.resultItem}>
      <a
        href={result.url}
        className={styles.resultTitle}
        target="_blank"
        rel="noopener noreferrer"
      >
        {result.title || result.url}
      </a>
      <p className={styles.resultURL}>{result.url}</p>
      <p className={styles.resultSnippet}><div dangerouslySetInnerHTML={{__html: result.snippet}}></div></p>
    </div>
    );
};

export default ResultCard;