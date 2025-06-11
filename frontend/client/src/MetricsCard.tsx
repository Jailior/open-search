import React from "react";
import styles from "./SearchPage.module.css"
import { FaSpider, FaRegClone, FaGlobe, FaSearch, FaBandAid, FaStar } from "react-icons/fa";

// Standard interface for metrics instance
export interface Metrics {
    pages_crawled: number[];
    queue_size: number[];
    page_errs: number;
    pages_skipped_lang: number;
    duplicates_avoided: number;
    number_of_searches: number;
}

interface Props {
    metrics: Metrics;
}

// Metrics card component, requires metrics instance
const MetricsCard: React.FC<Props> = ({ metrics }) => {
  // Get latest figures in list
  const latestCrawled =
    metrics.pages_crawled[metrics.pages_crawled.length - 1] || 0;
  const latestQueueSize =
    metrics.queue_size[metrics.queue_size.length - 1] || 0;

  return (
    <div className={styles.metricsItem}>
      <h1 className="text-4xl font-extrabold text-green-reseda mb-2">
        {metrics.number_of_searches.toLocaleString()}
      </h1>
      <h3 className="text-lg text-gray-600 mb-6">searches made on OpenSearch</h3>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-5 text-sm text-gray-700">
        <div className={styles.factText}>
          <FaSpider className="inline text-black-eerie mr-2" />  
          <strong>Total Pages Crawled:</strong> {latestCrawled.toLocaleString()}
        </div>
        <div className={styles.factText}>
           <FaSearch className="inline text-black-eerie mr-2" /> 
          <strong>Latest Queue Size:</strong> {latestQueueSize.toLocaleString()}
        </div>
        <div className={styles.factText}>
          <FaRegClone className="inline text-black-eerie mr-2" /> 
          <strong>Duplicates Avoided:</strong> {metrics.duplicates_avoided.toLocaleString()}
        </div>
        <div className={styles.factText}>
          <FaGlobe className="inline text-black-eerie mr-2" /> 
          <strong>Pages Skipped due to Lang:</strong> {metrics.pages_skipped_lang.toLocaleString()}
        </div>
        <div className={styles.factText}>
          <FaBandAid className="inline text-black-eerie mr-2" /> 
          <strong>Pages Skipped due to Errors:</strong> {metrics.page_errs.toLocaleString()}
        </div>
        <div className={styles.factText}>
          <FaStar className="inline text-yellow-500 mr-2" /> 
          <strong>Support the project on GitHub!</strong>
        </div>
      </div>
    </div>
  );
};

export default MetricsCard;