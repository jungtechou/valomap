import { renderHook, act } from '@testing-library/react-hooks';
import axios from 'axios';
import { useValorantAPI } from '../useValorantAPI';

// Mock axios
jest.mock('axios', () => ({
  create: jest.fn(() => ({
    get: jest.fn(),
    interceptors: {
      response: {
        use: jest.fn()
      }
    }
  }))
}));

describe('useValorantAPI', () => {
  let mockApiClient;
  let consoleSpy;

  beforeEach(() => {
    // Clear all mocks
    jest.clearAllMocks();

    // Setup the axios mock
    mockApiClient = axios.create();

    // Spy on console.error to verify error logging
    consoleSpy = jest.spyOn(console, 'error').mockImplementation(() => {});
  });

  afterEach(() => {
    consoleSpy.mockRestore();
  });

  test('should initialize with correct default values', () => {
    const { result } = renderHook(() => useValorantAPI());

    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBe(null);
    expect(typeof result.current.getRandomMap).toBe('function');
    expect(typeof result.current.clearCache).toBe('function');
    expect(typeof result.current.clearCacheForParams).toBe('function');
  });

  test('should handle successful API request', async () => {
    const mockData = { id: 'map1', name: 'Haven' };
    mockApiClient.get.mockResolvedValueOnce({ data: mockData });

    const { result, waitForNextUpdate } = renderHook(() => useValorantAPI());

    let promise;
    act(() => {
      promise = result.current.getRandomMap();
    });

    expect(result.current.loading).toBe(true);

    await waitForNextUpdate();

    // Resolve the promise
    const data = await promise;

    expect(data).toEqual(mockData);
    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBe(null);
    expect(mockApiClient.get).toHaveBeenCalledWith('/roulette');
  });

  test('should handle API error with response', async () => {
    const errorResponse = {
      response: {
        status: 500,
        statusText: 'Internal Server Error',
        data: { message: 'Server error' }
      }
    };

    mockApiClient.get.mockRejectedValueOnce(errorResponse);

    const { result, waitForNextUpdate } = renderHook(() => useValorantAPI());

    let error;
    act(() => {
      result.current.getRandomMap().catch(e => { error = e; });
    });

    expect(result.current.loading).toBe(true);

    await waitForNextUpdate();

    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBe('Server error');
    expect(consoleSpy).toHaveBeenCalled();
  });

  test('should handle network errors', async () => {
    const networkError = { request: {}, message: 'Network Error' };

    mockApiClient.get.mockRejectedValueOnce(networkError);

    const { result, waitForNextUpdate } = renderHook(() => useValorantAPI());

    let error;
    act(() => {
      result.current.getRandomMap().catch(e => { error = e; });
    });

    await waitForNextUpdate();

    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBe('No response from server. Please check your connection.');
  });

  test('should use cache for repeated requests', async () => {
    const mockData = { id: 'map1', name: 'Haven' };
    mockApiClient.get.mockResolvedValueOnce({ data: mockData });

    const { result, waitForNextUpdate } = renderHook(() => useValorantAPI());

    // First request
    let promise;
    act(() => {
      promise = result.current.getRandomMap();
    });

    await waitForNextUpdate();
    await promise;

    // Reset the mock to verify it's not called again
    mockApiClient.get.mockClear();

    // Second request with same params
    let secondPromise;
    act(() => {
      secondPromise = result.current.getRandomMap();
    });

    const secondData = await secondPromise;

    expect(secondData).toEqual(mockData);
    expect(mockApiClient.get).not.toHaveBeenCalled(); // Should use cache
  });

  test('should clear the cache', async () => {
    const mockData = { id: 'map1', name: 'Haven' };
    mockApiClient.get.mockResolvedValue({ data: mockData });

    const { result, waitForNextUpdate } = renderHook(() => useValorantAPI());

    // First request
    let promise;
    act(() => {
      promise = result.current.getRandomMap();
    });

    await waitForNextUpdate();
    await promise;

    // Clear cache
    act(() => {
      result.current.clearCache();
    });

    // Second request should call API again
    mockApiClient.get.mockClear();

    act(() => {
      promise = result.current.getRandomMap();
    });

    await waitForNextUpdate();

    expect(mockApiClient.get).toHaveBeenCalled();
  });

  test('should selectively clear cache for specific parameters', async () => {
    const mockData1 = { id: 'map1', name: 'Haven' };
    const mockData2 = { id: 'map2', name: 'Bind' };

    mockApiClient.get.mockResolvedValueOnce({ data: mockData1 })
                   .mockResolvedValueOnce({ data: mockData2 });

    const { result, waitForNextUpdate } = renderHook(() => useValorantAPI());

    // Request with standard maps
    act(() => {
      result.current.getRandomMap(true, []);
    });

    await waitForNextUpdate();

    // Request with non-standard maps
    mockApiClient.get.mockClear();

    act(() => {
      result.current.getRandomMap(false, []);
    });

    await waitForNextUpdate();

    // Clear cache for standard maps only
    act(() => {
      result.current.clearCacheForParams(true, []);
    });

    // Request standard maps again - should call API
    mockApiClient.get.mockClear();

    act(() => {
      result.current.getRandomMap(true, []);
    });

    await waitForNextUpdate();

    expect(mockApiClient.get).toHaveBeenCalled();

    // Request non-standard maps - should use cache
    mockApiClient.get.mockClear();

    act(() => {
      result.current.getRandomMap(false, []);
    });

    expect(mockApiClient.get).not.toHaveBeenCalled();
  });
});
