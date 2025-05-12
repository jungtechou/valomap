import { useState, useCallback, useRef, useMemo } from "react";
import axios from "axios";
import debounce from "lodash.debounce";

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
        // Use setTimeout to prevent blocking the main thread
        setTimeout(() => {
          setLoading(false);
        }, 0);
        return cache.current[cacheKey];
      }

      const response = await apiClient.get(url);

      // Update cache in a non-blocking way
      setTimeout(() => {
        cache.current[cacheKey] = response.data;
        setLoading(false);
      }, 0);

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
    // Start building the endpoint with query parameters
    let endpoint = '/roulette';
    const params = new URLSearchParams();

    // Add standard filter if needed
    if (standardOnly) {
      params.append('standard', 'true');
    }

    // Add banned maps if any - more efficient than building strings manually
    if (bannedMaps.length > 0) {
      bannedMaps.forEach(mapId => {
        params.append('banned', mapId);
      });
    }

    // Construct the final endpoint with query parameters
    if (params.toString()) {
      endpoint += '?' + params.toString();
    }

    // Generate a unique cache key based on the parameters
    const cacheKey = generateCacheKey(standardOnly, bannedMaps);

    return makeRequest(endpoint, cacheKey);
  }, [makeRequest, generateCacheKey]);

  // Create a debounced version of the function to prevent too many requests when rapidly toggling maps
  const getRandomMap = useMemo(() => {
    const debouncedFn = debounce(getRandomMapInternal, 100, { leading: true, trailing: false });

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
