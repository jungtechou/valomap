import React from 'react';
import { render, screen, act } from '@testing-library/react';
import MapCard from '../MapCard';
import * as imageUtils from '../../utils/imageUtils';

// Mock the imageUtils module
jest.mock('../../utils/imageUtils', () => ({
  ensureValidImageUrl: jest.fn(url => url ? `processed-${url}` : null),
}));

describe('MapCard', () => {
  // Mock Image implementation
  const originalImage = global.Image;

  beforeEach(() => {
    // Mock the Image constructor
    global.Image = class {
      constructor() {
        setTimeout(() => {
          this.onload && this.onload();
        }, 0);
      }
    };
  });

  afterEach(() => {
    global.Image = originalImage;
    jest.clearAllMocks();
  });

  test('renders empty state when no map is provided', () => {
    render(<MapCard map={null} />);

    expect(screen.getByText(/Click the button to randomly select a VALORANT map/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/No map selected/i)).toBeInTheDocument();
  });

  test('renders map details when map is provided', async () => {
    const mockMap = {
      uuid: 'test-uuid',
      displayName: 'Test Map',
      splash: 'test-splash-url',
    };

    // Render component
    render(<MapCard map={mockMap} />);

    // Check map name is rendered
    expect(screen.getByText('Test Map')).toBeInTheDocument();

    // Check image processing
    expect(imageUtils.ensureValidImageUrl).toHaveBeenCalledWith('test-splash-url');
  });

  test('uses displayIcon when splash is not available', () => {
    const mockMap = {
      uuid: 'test-uuid',
      displayName: 'Test Map',
      displayIcon: 'test-icon-url',
    };

    render(<MapCard map={mockMap} />);

    expect(imageUtils.ensureValidImageUrl).toHaveBeenCalledWith('test-icon-url');
  });

  test('uses fallback image when no images are available', () => {
    const mockMap = {
      uuid: 'test-uuid',
      displayName: 'Test Map',
    };

    render(<MapCard map={mockMap} />);

    expect(imageUtils.ensureValidImageUrl).toHaveBeenCalledWith(
      'https://via.placeholder.com/700x300?text=Map+Image+Unavailable'
    );
  });

  test('handles image loading state', async () => {
    // Create a controlled image loading mock
    let imageLoadCallback = null;

    global.Image = class {
      constructor() {
        this.onload = null;
      }

      set src(value) {
        // Store the callback for external control
        imageLoadCallback = () => this.onload && this.onload();
      }
    };

    const mockMap = {
      uuid: 'test-uuid',
      displayName: 'Test Map',
      splash: 'test-splash-url',
    };

    // Render with loading image
    const { rerender } = render(<MapCard map={mockMap} />);

    // Get the image element
    const imageElement = screen.getByLabelText(`${mockMap.displayName} map image`);

    // Initially, the image should be transparent
    expect(imageElement).toHaveStyle({ opacity: '0' });

    // Simulate image loaded
    await act(async () => {
      imageLoadCallback();
    });

    // Re-render to see updated state
    rerender(<MapCard map={mockMap} />);

    // After loading, the image should be visible
    expect(screen.getByLabelText(`${mockMap.displayName} map image`)).toHaveStyle({ opacity: '1' });
  });

  test('resets loading state when map changes', () => {
    // Create a controlled image loading mock
    let imageLoadCallback = null;

    global.Image = class {
      constructor() {
        this.onload = null;
      }

      set src(value) {
        imageLoadCallback = () => this.onload && this.onload();
      }
    };

    const mockMap1 = {
      uuid: 'map-1',
      displayName: 'First Map',
      splash: 'splash-1',
    };

    const mockMap2 = {
      uuid: 'map-2',
      displayName: 'Second Map',
      splash: 'splash-2',
    };

    // Render with first map
    const { rerender } = render(<MapCard map={mockMap1} />);

    // Simulate image loaded
    act(() => {
      imageLoadCallback();
    });

    // Change to second map
    rerender(<MapCard map={mockMap2} />);

    // Image should be transparent again
    expect(screen.getByLabelText(`${mockMap2.displayName} map image`)).toHaveStyle({ opacity: '0' });
  });
});
