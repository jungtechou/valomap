import React, { useState, useEffect, useCallback, useMemo } from 'react';
import styled from 'styled-components';
import { motion, AnimatePresence } from 'framer-motion';
import { FaBan, FaTimes } from 'react-icons/fa';
import axios from 'axios';
import { useInView } from 'react-intersection-observer';
import { ensureValidImageUrl } from '../utils/imageUtils';

const BanSelectionContainer = styled(motion.div)`
  width: 100%;
  max-width: 800px;
  margin: 2rem 0;
`;

const Title = styled.h3`
  font-size: 1.2rem;
  font-weight: 600;
  color: var(--valorant-light-gray);
  margin-bottom: 1rem;
  display: flex;
  align-items: center;

  svg {
    color: var(--valorant-red);
    margin-right: 0.5rem;
  }
`;

const MapGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 1rem;
  margin-top: 1rem;

  @media (max-width: 480px) {
    grid-template-columns: repeat(2, 1fr);
  }
`;

const MapItem = styled.div`
  position: relative;
  border-radius: 6px;
  overflow: hidden;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease;
  background-color: rgba(15, 25, 35, 0.6);
  border: 2px solid ${props => props.isBanned ? 'var(--valorant-red)' : 'rgba(255, 255, 255, 0.1)'};
  will-change: transform, box-shadow, border-color;

  &:hover {
    transform: translateY(-3px);
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
    border-color: ${props => props.isBanned ? 'var(--valorant-red)' : 'rgba(255, 255, 255, 0.3)'};
  }
`;

const MapImage = styled.div`
  width: 100%;
  height: 80px;
  background-image: ${props => props.loaded ? `url(${props.src})` : 'none'};
  background-color: rgba(15, 25, 35, 0.8);
  background-size: cover;
  background-position: center;
  position: relative;

  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(0, 0, 0, ${props => props.isBanned ? '0.7' : '0.2'});
    z-index: 1;
  }
`;

const MapName = styled.div`
  padding: 0.5rem;
  text-align: center;
  font-size: 0.9rem;
  font-weight: 500;
  color: ${props => props.isBanned ? 'var(--valorant-red)' : 'var(--valorant-white)'};
`;

const BanOverlay = styled.div`
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--valorant-red);
  font-size: 2rem;
  z-index: 2;
  opacity: ${props => props.visible ? 1 : 0};
  transition: opacity 0.2s ease;
  pointer-events: none;
`;

const BannedMapsCount = styled.div`
  margin-top: 1rem;
  font-size: 0.9rem;
  color: var(--valorant-light-gray);

  span {
    color: var(--valorant-red);
    font-weight: 600;
  }
`;

const LoadingContainer = styled.div`
  width: 100%;
  padding: 2rem;
  text-align: center;
  color: var(--valorant-light-gray);
`;

const ErrorContainer = styled.div`
  width: 100%;
  padding: 1rem;
  text-align: center;
  color: var(--valorant-red);
  background-color: rgba(255, 70, 85, 0.1);
  border-radius: 4px;
  margin-top: 1rem;
`;

// Memoized map item component to prevent unnecessary re-renders
const MapItemMemo = React.memo(({ map, isBanned, toggleMapBan }) => {
  const [imageLoaded, setImageLoaded] = useState(false);
  const [ref, inView] = useInView({
    triggerOnce: true,
    rootMargin: '200px 0px',
  });

  // Process the image URLs
  const imageUrl = ensureValidImageUrl(map.splash || map.displayIcon);

  // Preload image when component comes into view
  useEffect(() => {
    if (inView && !imageLoaded && imageUrl) {
      const img = new Image();
      img.src = imageUrl;
      img.onload = () => setImageLoaded(true);
    }
  }, [inView, imageLoaded, imageUrl]);

  return (
    <MapItem
      ref={ref}
      isBanned={isBanned}
      onClick={() => toggleMapBan(map.uuid)}
      role="checkbox"
      aria-checked={isBanned}
      aria-label={`${map.displayName} ${isBanned ? 'banned' : 'not banned'}`}
    >
      <MapImage
        src={imageUrl}
        isBanned={isBanned}
        loaded={inView && imageLoaded}
        aria-hidden="true"
      />
      <MapName isBanned={isBanned}>{map.displayName}</MapName>
      <BanOverlay visible={isBanned}>
        <FaBan />
      </BanOverlay>
    </MapItem>
  );
});

const MapBanSelection = ({ bannedMaps = [], onBannedMapsChange }) => {
  const [maps, setMaps] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  // Use useCallback to prevent unnecessary rebuilds of this function
  const fetchMaps = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await axios.get('/map/all');
      setMaps(response.data);
    } catch (err) {
      console.error('Failed to fetch maps:', err);
      setError('Failed to load maps. Please try again later.');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchMaps();
  }, [fetchMaps]);

  // Memoize the toggle function to prevent recreation on each render
  const toggleMapBan = useCallback((mapId) => {
    let newBannedMaps;

    if (bannedMaps.includes(mapId)) {
      // Remove the map from banned list
      newBannedMaps = bannedMaps.filter(id => id !== mapId);
    } else {
      // Add the map to banned list
      newBannedMaps = [...bannedMaps, mapId];
    }

    if (onBannedMapsChange) {
      onBannedMapsChange(newBannedMaps);
    }
  }, [bannedMaps, onBannedMapsChange]);

  // Memoize the banned maps set for faster lookups
  const bannedMapsSet = useMemo(() => {
    return new Set(bannedMaps);
  }, [bannedMaps]);

  if (loading) {
    return <LoadingContainer>Loading maps...</LoadingContainer>;
  }

  if (error) {
    return <ErrorContainer>{error}</ErrorContainer>;
  }

  return (
    <BanSelectionContainer
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
    >
      <Title>
        <FaBan /> Ban Maps
      </Title>
      <div>Select maps you want to exclude from the roulette:</div>

      <MapGrid>
        {maps.map(map => (
          <MapItemMemo
            key={map.uuid}
            map={map}
            isBanned={bannedMapsSet.has(map.uuid)}
            toggleMapBan={toggleMapBan}
          />
        ))}
      </MapGrid>

      <BannedMapsCount>
        <span>{bannedMaps.length}</span> map{bannedMaps.length !== 1 ? 's' : ''} banned
      </BannedMapsCount>
    </BanSelectionContainer>
  );
};

export default React.memo(MapBanSelection);
