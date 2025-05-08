import { useState, useCallback, useRef } from "react";
import axios from "axios";

const API_BASE_URL = "/map";

// Create axios instance with defaults
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
  }
});

export const useValorantAPI = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // Use a ref for caching to persist across renders
  const cache = useRef({
    standard: null,
    all: null
  });

  // Create a single function to handle API requests
  const makeRequest = useCallback(async (url, cacheKey) => {
    setLoading(true);
    setError(null);

    try {
      // Check cache first
      if (cache.current[cacheKey]) {
        setLoading(false);
        return cache.current[cacheKey];
      }

      const response = await apiClient.get(url);

      // Update cache
      cache.current[cacheKey] = response.data;

      setLoading(false);
      return response.data;
    } catch (err) {
      // Detailed error handling
      let errorMessage = 'Failed to fetch map data';

      if (err.response) {
        // The request was made and the server responded with an error
        errorMessage = err.response.data?.message || `Error ${err.response.status}: ${err.response.statusText}`;
      } else if (err.request) {
        // The request was made but no response was received
        errorMessage = 'No response from server. Please check your connection.';
      }

      setError(errorMessage);
      setLoading(false);
      throw new Error(errorMessage);
    }
  }, []);

  // Get a random map with optional standard filter
  const getRandomMap = useCallback(async (standardOnly = false) => {
    const endpoint = standardOnly ? '/roulette?standard=true' : '/roulette';
    const cacheKey = standardOnly ? 'standard' : 'all';

    return makeRequest(endpoint, cacheKey);
  }, [makeRequest]);

  // Clear cached data - useful when implementing manual refresh
  const clearCache = useCallback(() => {
    cache.current = {
      standard: null,
      all: null
    };
  }, []);

  return {
    loading,
    error,
    getRandomMap,
    clearCache
  };
};
