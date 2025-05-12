/**
 * Utility functions for handling image URLs
 */

/**
 * Ensure cache URLs are properly formatted with the correct domain
 * This resolves the issue with redirected cache URLs
 *
 * @param {string} url - The image URL to process
 * @returns {string} - The properly formatted URL
 */
export const ensureValidImageUrl = (url) => {
  if (!url) return null;

  // If it's already an absolute URL with a different domain, return as is
  if (url.startsWith('http') && !url.includes('valomap.com')) {
    return url;
  }

  // If it's a cache URL, ensure it has the correct domain
  if (url.includes('/api/cache/')) {
    // Extract just the filename
    const filename = url.split('/').pop();
    // Construct the fully qualified URL with the current domain
    return `${window.location.protocol}//${window.location.host}/api/cache/${filename}`;
  }

  return url;
};

/**
 * Calculate optimal image size based on device
 *
 * @param {number} baseWidth - The base width of the image
 * @returns {number} - The optimal width for the current device
 */
export const getOptimalImageSize = (baseWidth = 300) => {
  const dpr = window.devicePixelRatio || 1;
  return Math.round(baseWidth * Math.min(dpr, 2)); // Cap at 2x for performance
};

/**
 * Preload an image and return a promise
 *
 * @param {string} src - The image URL to preload
 * @returns {Promise} - Resolves when image is loaded, rejects on error
 */
export const preloadImage = (src) => {
  if (!src) return Promise.reject(new Error('No image URL provided'));

  return new Promise((resolve, reject) => {
    const img = new Image();
    img.decoding = 'async';
    img.onload = () => resolve(img);
    img.onerror = () => reject(new Error(`Failed to load image: ${src}`));
    img.src = src;
  });
};
