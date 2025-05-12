import React from 'react';
import { render, screen, fireEvent, act } from '@testing-library/react';
import axios from 'axios';
import MapBanSelection from '../MapBanSelection';
import { useInView } from 'react-intersection-observer';

// Mock dependencies
jest.mock('axios');
jest.mock('react-intersection-observer');
jest.mock('../../utils/imageUtils', () => ({
  ensureValidImageUrl: jest.fn(url => url ? `processed-${url}` : null),
}));

describe('MapBanSelection', () => {
  const mockMaps = [
    { uuid: 'map1', displayName: 'Haven', splash: 'haven.jpg' },
    { uuid: 'map2', displayName: 'Bind', splash: 'bind.jpg' },
    { uuid: 'map3', displayName: 'Split', splash: 'split.jpg' }
  ];

  beforeEach(() => {
    // Mock axios.get to return mock maps
    axios.get.mockResolvedValue({ data: mockMaps });

    // Mock useInView to return always in view
    useInView.mockReturnValue([
      (element) => {}, // ref function
      true,           // inView
      {}              // entry
    ]);

    // Mock Image constructor and loading behavior
    global.Image = class {
      constructor() {
        this.onload = null;
        this.onerror = null;
      }

      set src(value) {
        // Simulate successful image load
        setTimeout(() => {
          this.onload && this.onload();
        }, 0);
      }
    };
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  test('renders loading state initially', () => {
    render(<MapBanSelection bannedMaps={[]} onBannedMapsChange={() => {}} />);

    expect(screen.getByText('Loading maps...')).toBeInTheDocument();
  });

  test('renders maps after loading', async () => {
    await act(async () => {
      render(<MapBanSelection bannedMaps={[]} onBannedMapsChange={() => {}} />);
    });

    // Check if all map names are rendered
    expect(screen.getByText('Haven')).toBeInTheDocument();
    expect(screen.getByText('Bind')).toBeInTheDocument();
    expect(screen.getByText('Split')).toBeInTheDocument();

    // Check banned count
    expect(screen.getByText('0')).toBeInTheDocument();
    expect(screen.getByText('maps banned')).toBeInTheDocument();
  });

  test('handles error state', async () => {
    // Mock API error
    axios.get.mockRejectedValueOnce(new Error('API error'));

    await act(async () => {
      render(<MapBanSelection bannedMaps={[]} onBannedMapsChange={() => {}} />);
    });

    expect(screen.getByText('Failed to load maps. Please try again later.')).toBeInTheDocument();
  });

  test('toggles map ban on click', async () => {
    const handleBannedMapsChange = jest.fn();

    await act(async () => {
      render(
        <MapBanSelection
          bannedMaps={[]}
          onBannedMapsChange={handleBannedMapsChange}
        />
      );
    });

    // Click on the first map
    await act(async () => {
      fireEvent.click(screen.getByText('Haven'));
    });

    // Check if callback was called with the correct map ID
    expect(handleBannedMapsChange).toHaveBeenCalledWith(['map1']);
  });

  test('unbans a previously banned map', async () => {
    const handleBannedMapsChange = jest.fn();

    await act(async () => {
      render(
        <MapBanSelection
          bannedMaps={['map1']}
          onBannedMapsChange={handleBannedMapsChange}
        />
      );
    });

    // Click on the already banned map
    await act(async () => {
      fireEvent.click(screen.getByText('Haven'));
    });

    // Check if callback was called with empty array
    expect(handleBannedMapsChange).toHaveBeenCalledWith([]);
  });

  test('renders correct banned map count', async () => {
    await act(async () => {
      render(
        <MapBanSelection
          bannedMaps={['map1', 'map2']}
          onBannedMapsChange={() => {}}
        />
      );
    });

    expect(screen.getByText('2')).toBeInTheDocument();
    expect(screen.getByText('maps banned')).toBeInTheDocument();
  });

  test('renders singular text for one banned map', async () => {
    await act(async () => {
      render(
        <MapBanSelection
          bannedMaps={['map1']}
          onBannedMapsChange={() => {}}
        />
      );
    });

    expect(screen.getByText('1')).toBeInTheDocument();
    expect(screen.getByText('map banned')).toBeInTheDocument();
  });

  test('lazy loads images when they come into view', async () => {
    // Mock image loading with controlled behavior
    let imageLoadCallbacks = [];
    global.Image = class {
      constructor() {
        this.onload = null;
        this.loading = '';
        this.decoding = '';
      }

      set src(value) {
        imageLoadCallbacks.push(() => this.onload && this.onload());
      }
    };

    // Control the in-view status
    useInView.mockReturnValue([
      (element) => {}, // ref function
      true,           // inView - element is in view
      {}              // entry
    ]);

    await act(async () => {
      render(<MapBanSelection bannedMaps={[]} onBannedMapsChange={() => {}} />);
    });

    // Trigger image load callbacks
    await act(async () => {
      imageLoadCallbacks.forEach(callback => callback());
    });

    // Check for proper attributes on images
    const images = document.querySelectorAll('[aria-hidden="true"]');
    expect(images.length).toBe(3); // All 3 maps have images
  });
});
