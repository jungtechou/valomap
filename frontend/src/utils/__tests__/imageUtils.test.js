import { ensureValidImageUrl, getOptimalImageSize, preloadImage } from '../imageUtils';

// Mock window properties
const originalLocation = window.location;

beforeAll(() => {
  delete window.location;
  window.location = {
    protocol: 'https:',
    host: 'test.valomap.com'
  };
});

afterAll(() => {
  window.location = originalLocation;
});

describe('ensureValidImageUrl', () => {
  test('returns null for null input', () => {
    expect(ensureValidImageUrl(null)).toBeNull();
  });

  test('returns the same URL for external URLs', () => {
    const url = 'https://example.com/image.jpg';
    expect(ensureValidImageUrl(url)).toBe(url);
  });

  test('formats cache URLs correctly', () => {
    const url = '/api/cache/map_image.jpg';
    expect(ensureValidImageUrl(url)).toBe('https://test.valomap.com/api/cache/map_image.jpg');
  });

  test('handles regular URLs correctly', () => {
    const url = '/local/image.jpg';
    expect(ensureValidImageUrl(url)).toBe(url);
  });
});

describe('getOptimalImageSize', () => {
  test('calculates optimal size based on DPR', () => {
    // Mock devicePixelRatio
    Object.defineProperty(window, 'devicePixelRatio', {
      writable: true,
      value: 2
    });

    expect(getOptimalImageSize(300)).toBe(600);
  });

  test('caps at 2x even for higher DPR', () => {
    // Mock devicePixelRatio
    Object.defineProperty(window, 'devicePixelRatio', {
      writable: true,
      value: 3
    });

    expect(getOptimalImageSize(300)).toBe(600);
  });

  test('uses default base width when not provided', () => {
    Object.defineProperty(window, 'devicePixelRatio', {
      writable: true,
      value: 1.5
    });

    expect(getOptimalImageSize()).toBe(450);
  });
});

describe('preloadImage', () => {
  test('rejects for null input', async () => {
    await expect(preloadImage(null)).rejects.toThrow('No image URL provided');
  });

  test('resolves when image loads successfully', async () => {
    // Mock the Image constructor
    const mockImage = {
      onload: null,
      onerror: null,
      decoding: '',
      src: ''
    };

    global.Image = jest.fn(() => mockImage);

    const promise = preloadImage('https://test.com/image.jpg');

    // Simulate the image loading
    mockImage.onload();

    await expect(promise).resolves.toBe(mockImage);
    expect(mockImage.decoding).toBe('async');
    expect(mockImage.src).toBe('https://test.com/image.jpg');
  });

  test('rejects when image fails to load', async () => {
    // Mock the Image constructor
    const mockImage = {
      onload: null,
      onerror: null,
      decoding: '',
      src: ''
    };

    global.Image = jest.fn(() => mockImage);

    const promise = preloadImage('https://test.com/image.jpg');

    // Simulate the image error
    mockImage.onerror();

    await expect(promise).rejects.toThrow('Failed to load image');
  });
});
