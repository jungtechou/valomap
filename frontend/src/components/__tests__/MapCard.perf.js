import React from 'react';
import { render, unmountComponentAtNode } from 'react-dom';
import MapCard from '../MapCard';

// Mock the imageUtils module
jest.mock('../../utils/imageUtils', () => ({
  ensureValidImageUrl: jest.fn(url => url ? `processed-${url}` : null),
}));

describe('MapCard Performance', () => {
  let container = null;

  beforeEach(() => {
    // Setup a DOM element as the render target
    container = document.createElement('div');
    document.body.appendChild(container);

    // Mock Image constructor behavior
    global.Image = class {
      constructor() {
        setTimeout(() => {
          this.onload && this.onload();
        }, 5);
      }
    };

    // Mock performance API if not available
    if (!window.performance) {
      window.performance = {
        mark: jest.fn(),
        measure: jest.fn(),
        getEntriesByName: jest.fn().mockReturnValue([{ duration: 15 }]),
        clearMarks: jest.fn(),
        clearMeasures: jest.fn()
      };
    }
  });

  afterEach(() => {
    // Cleanup on exit
    unmountComponentAtNode(container);
    container.remove();
    container = null;
  });

  // Test initial render time
  it('renders initial empty state quickly', () => {
    // Start measuring
    performance.mark('start-render');

    // Render component
    render(<MapCard map={null} />, container);

    // End measuring
    performance.mark('end-render');
    performance.measure('render-time', 'start-render', 'end-render');

    // Get results
    const measurements = performance.getEntriesByName('render-time');
    const duration = measurements[0].duration;

    // Clean up
    performance.clearMarks();
    performance.clearMeasures();

    // Log and assert
    console.log(`Empty MapCard render time: ${duration}ms`);
    expect(duration).toBeLessThan(50); // Should render in less than 50ms
  });

  // Test render with map data
  it('renders map data efficiently', () => {
    const mockMap = {
      uuid: 'test-uuid',
      displayName: 'Test Map',
      splash: 'test-splash-url',
    };

    // Start measuring
    performance.mark('start-render-data');

    // Render component with data
    render(<MapCard map={mockMap} />, container);

    // End measuring
    performance.mark('end-render-data');
    performance.measure('render-time-data', 'start-render-data', 'end-render-data');

    // Get results
    const measurements = performance.getEntriesByName('render-time-data');
    const duration = measurements[0].duration;

    // Clean up
    performance.clearMarks();
    performance.clearMeasures();

    // Log and assert
    console.log(`MapCard with data render time: ${duration}ms`);
    expect(duration).toBeLessThan(100); // Should render in less than 100ms with data
  });

  // Test re-render performance with new map
  it('re-renders efficiently when map changes', () => {
    const mockMap1 = {
      uuid: 'map1',
      displayName: 'First Map',
      splash: 'splash-1',
    };

    const mockMap2 = {
      uuid: 'map2',
      displayName: 'Second Map',
      splash: 'splash-2',
    };

    // Initial render
    render(<MapCard map={mockMap1} />, container);

    // Start measuring for re-render
    performance.mark('start-rerender');

    // Re-render with new map
    render(<MapCard map={mockMap2} />, container);

    // End measuring
    performance.mark('end-rerender');
    performance.measure('rerender-time', 'start-rerender', 'end-rerender');

    // Get results
    const measurements = performance.getEntriesByName('rerender-time');
    const duration = measurements[0].duration;

    // Clean up
    performance.clearMarks();
    performance.clearMeasures();

    // Log and assert
    console.log(`MapCard re-render time: ${duration}ms`);
    expect(duration).toBeLessThan(80); // Re-render should be faster than initial render
  });
});
