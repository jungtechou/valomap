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
