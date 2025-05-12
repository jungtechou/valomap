import { useState, useCallback, useRef, useMemo } from "react";
import axios from "axios";
import debounce from "lodash.debounce";

const API_BASE_URL = "/map";

// Create axios instance with improved defaults
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 8000, // Reduced timeout for faster error detection
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
  }
});

// Add response interceptor for global error handling
apiClient.interceptors.response.use(
  response => response,
  error => {
    // Log error for debugging
    console.error('API request failed:', error);
    return Promise.reject(error);
  }
);

export const useValorantAPI = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // Use a ref for caching to persist across renders
  const cache = useRef({});

  // Function to generate a cache key
  const generateCacheKey = useCallback((standardOnly, bannedMaps) => {
    return `${standardOnly ? 'standard' : 'all'}_${bannedMaps.sort().join('_')}`;
  }, []);

  // Create a single function to handle API requests with proper error handling
  const makeRequest = useCallback(async (url, cacheKey) => {
    setLoading(true);
    setError(null);

    try {
      // Check cache first
      if (cache.current[cacheKey]) {
        // Return immediately from cache to prevent loading flicker
        setLoading(false);
        return cache.current[cacheKey];
      }

      const response = await apiClient.get(url);

      // Add to cache immediately
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

  // Get a random map with optional standard filter and banned maps
  const getRandomMapInternal = useCallback(async (standardOnly = false, bannedMaps = []) => {
    // Build query params more efficiently
    const params = new URLSearchParams();

    if (standardOnly) {
      params.append('standard', 'true');
    }

    if (bannedMaps.length > 0) {
      bannedMaps.forEach(mapId => {
        params.append('banned', mapId);
      });
    }

    const endpoint = `/roulette${params.toString() ? '?' + params.toString() : ''}`;
    const cacheKey = generateCacheKey(standardOnly, bannedMaps);

    return makeRequest(endpoint, cacheKey);
  }, [makeRequest, generateCacheKey]);

  // Create a debounced version of the function with shorter delay
  const getRandomMap = useMemo(() => {
    const debouncedFn = debounce(getRandomMapInternal, 50, { leading: true, trailing: false });

    // Wrap in another function to preserve the promise interface
    return async (standardOnly = false, bannedMaps = []) => {
      return debouncedFn(standardOnly, bannedMaps);
    };
  }, [getRandomMapInternal]);

  // Clear cached data - useful when implementing manual refresh
  const clearCache = useCallback(() => {
    cache.current = {};
  }, []);

  // Selectively clear cache for specific params
  const clearCacheForParams = useCallback((standardOnly, bannedMaps) => {
    const cacheKey = generateCacheKey(standardOnly, bannedMaps);
    if (cache.current[cacheKey]) {
      delete cache.current[cacheKey];
    }
  }, [generateCacheKey]);

  return {
    loading,
    error,
    getRandomMap,
    clearCache,
    clearCacheForParams
  };
};
