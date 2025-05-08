import React, { useState, useCallback } from 'react';
import styled from 'styled-components';
import { motion, AnimatePresence } from 'framer-motion';
import { FaDice, FaDiceD6 } from 'react-icons/fa';
import { useValorantAPI } from '../hooks/useValorantAPI';
import MapCard from './MapCard';
import Button from './Button';

const RouletteContainer = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
  max-width: 800px;
  margin: 0 auto;
`;

const Title = styled(motion.h1)`
  font-size: 2.5rem;
  font-weight: 700;
  color: var(--valorant-white);
  text-align: center;
  margin-bottom: 1.5rem;

  span {
    color: var(--valorant-red);
  }

  @media (max-width: 768px) {
    font-size: 2rem;
  }
`;

const Subtitle = styled(motion.p)`
  font-size: 1.1rem;
  color: var(--valorant-light-gray);
  text-align: center;
  margin-bottom: 2.5rem;
  max-width: 600px;

  @media (max-width: 768px) {
    font-size: 0.9rem;
    margin-bottom: 2rem;
  }
`;

const ButtonsContainer = styled.div`
  display: flex;
  gap: 1rem;
  margin: 2rem 0;
  flex-wrap: wrap;
  justify-content: center;

  @media (max-width: 480px) {
    flex-direction: column;
    width: 100%;
  }
`;

const FilterToggle = styled.div`
  display: flex;
  align-items: center;
  margin-bottom: 1.5rem;

  label {
    color: var(--valorant-light-gray);
    margin-left: 10px;
    cursor: pointer;
    font-size: 0.9rem;
  }
`;

const Toggle = styled.div`
  position: relative;
  width: 50px;
  height: 24px;
  border-radius: 12px;
  background-color: ${props => props.checked ? 'var(--valorant-red)' : 'var(--valorant-dark-gray)'};
  cursor: pointer;
  transition: background-color 0.3s ease;

  &::after {
    content: '';
    position: absolute;
    top: 2px;
    left: ${props => props.checked ? '26px' : '2px'};
    width: 20px;
    height: 20px;
    border-radius: 50%;
    background-color: var(--valorant-white);
    transition: left 0.3s ease;
  }
`;

const LoadingIndicator = styled(motion.div)`
  width: 100%;
  display: flex;
  justify-content: center;
  margin: 1rem 0;

  svg {
    color: var(--valorant-red);
    font-size: 2rem;
  }
`;

const ErrorMessage = styled(motion.div)`
  background-color: rgba(255, 70, 85, 0.2);
  border: 1px solid var(--valorant-red);
  padding: 1rem;
  border-radius: 4px;
  margin: 1rem 0;
  color: var(--valorant-red);
  font-size: 0.9rem;
  width: 100%;
  text-align: center;
  display: flex;
  align-items: center;
  justify-content: center;
`;

const MapRoulette = () => {
  const [selectedMap, setSelectedMap] = useState(null);
  const [standardOnly, setStandardOnly] = useState(false);
  const { loading, error, getRandomMap } = useValorantAPI();

  const handleToggleStandard = useCallback(() => {
    setStandardOnly(prev => !prev);
  }, []);

  const handleRandomMap = useCallback(async () => {
    try {
      const map = await getRandomMap(standardOnly);
      setSelectedMap(map);
    } catch (error) {
      // Error is handled in the useValorantAPI hook
      console.error('Failed to get random map:', error);
    }
  }, [getRandomMap, standardOnly]);

  const handleReset = useCallback(() => {
    setSelectedMap(null);
  }, []);

  // Animation variants
  const containerVariants = {
    hidden: { opacity: 0 },
    visible: {
      opacity: 1,
      transition: {
        staggerChildren: 0.2
      }
    }
  };

  const itemVariants = {
    hidden: { opacity: 0, y: 20 },
    visible: { opacity: 1, y: 0 }
  };

  return (
    <RouletteContainer
      as={motion.div}
      variants={containerVariants}
      initial="hidden"
      animate="visible"
    >
      <Title variants={itemVariants}>
        VALORANT <span>MAP ROULETTE</span>
      </Title>

      <Subtitle variants={itemVariants}>
        Randomly select a map for your next Valorant match.
        Use the toggle to filter only standard maps.
      </Subtitle>

      <FilterToggle as={motion.div} variants={itemVariants}>
        <Toggle
          checked={standardOnly}
          onClick={handleToggleStandard}
          aria-label={standardOnly ? "Show all maps" : "Show only standard maps"}
          role="switch"
          aria-checked={standardOnly}
        />
        <label onClick={handleToggleStandard}>
          Standard Maps Only
        </label>
      </FilterToggle>

      <AnimatePresence>
        {error && (
          <ErrorMessage
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0.9 }}
            transition={{ duration: 0.3 }}
          >
            {error}
          </ErrorMessage>
        )}
      </AnimatePresence>

      <MapCard map={selectedMap} />

      <ButtonsContainer as={motion.div} variants={itemVariants}>
        <Button
          onClick={handleRandomMap}
          disabled={loading}
          aria-label="Get random map"
        >
          {loading ? 'Selecting...' : 'Random Map'} <FaDice />
        </Button>

        <Button
          variant="secondary"
          onClick={handleReset}
          disabled={!selectedMap || loading}
          aria-label="Reset selection"
        >
          Reset <FaDiceD6 />
        </Button>
      </ButtonsContainer>

      <AnimatePresence>
        {loading && (
          <LoadingIndicator
            animate={{ rotate: 360 }}
            transition={{ repeat: Infinity, duration: 1, ease: "linear" }}
          >
            <FaDice />
          </LoadingIndicator>
        )}
      </AnimatePresence>
    </RouletteContainer>
  );
};

export default React.memo(MapRoulette);
