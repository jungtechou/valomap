import React from "react";
import styled from "styled-components";
import { motion, AnimatePresence } from "framer-motion";
import { ensureValidImageUrl } from "../utils/imageUtils";

const CardContainer = styled(motion.div)`
  width: 100%;
  max-width: 700px;
  border-radius: 10px;
  overflow: hidden;
  background-color: rgba(15, 25, 35, 0.8);
  border: 2px solid rgba(255, 70, 85, 0.3);
  box-shadow: 0 10px 20px rgba(0, 0, 0, 0.3);
  transition: all 0.3s ease;
  margin: 0 auto;

  @media (max-width: 768px) {
    max-width: 100%;
  }
`;

const MapImage = styled.div`
  width: 100%;
  height: 300px;
  background-image: url(${(props) => props.src});
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
    background-color: rgba(0, 0, 0, 0.2);
    z-index: 1;
  }

  &::after {
    content: "";
    position: absolute;
    bottom: 0;
    left: 0;
    width: 100%;
    height: 100px;
    background: linear-gradient(to top, rgba(15, 25, 35, 1), transparent);
    z-index: 2;
  }

  @media (max-width: 768px) {
    height: 200px;
  }
`;

const CardContent = styled.div`
  padding: 1.5rem;
`;

const MapName = styled.h2`
  font-size: 2rem;
  font-weight: 700;
  color: var(--valorant-white);
  margin-bottom: 1rem;
  display: flex;
  align-items: center;

  &::before {
    content: "";
    display: inline-block;
    width: 5px;
    height: 32px;
    background-color: var(--valorant-red);
    margin-right: 15px;
  }

  @media (max-width: 768px) {
    font-size: 1.5rem;
  }
`;

const MapDescription = styled.p`
  color: var(--valorant-light-gray);
  font-size: 1rem;
  line-height: 1.6;
  margin-bottom: 1rem;
`;

const MapCoordinates = styled.div`
  display: inline-block;
  padding: 0.5rem 1rem;
  background-color: rgba(255, 70, 85, 0.1);
  border-radius: 4px;
  font-size: 0.9rem;
  color: var(--valorant-light-gray);
  margin-top: 0.5rem;
`;

const NoMapMessage = styled.div`
  padding: 2rem;
  text-align: center;
  font-size: 1.2rem;
  color: var(--valorant-light-gray);
`;

const MapDetailSection = styled.section`
  margin-bottom: 1.5rem;
`;

const MapDetailHeading = styled.h3`
  font-size: 1rem;
  font-weight: 600;
  color: var(--valorant-red);
  margin-bottom: 0.5rem;
  text-transform: uppercase;
  letter-spacing: 1px;
`;

// Animation variants
const mapCardVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: {
    opacity: 1,
    y: 0,
    transition: {
      duration: 0.5,
      ease: "easeOut"
    }
  },
  exit: {
    opacity: 0,
    y: -20,
    transition: {
      duration: 0.3,
      ease: "easeIn"
    }
  }
};

const MapCard = ({ map }) => {
  if (!map) {
    return (
      <CardContainer
        aria-label="No map selected"
        initial="hidden"
        animate="visible"
        variants={mapCardVariants}
      >
        <NoMapMessage>
          Click the button to randomly select a VALORANT map.
        </NoMapMessage>
      </CardContainer>
    );
  }

  // Fallback image if splash or displayIcon is not available
  let mapImage = map.splash || map.displayIcon || "https://via.placeholder.com/700x300?text=Map+Image+Unavailable";

  // Ensure the image URL is valid
  mapImage = ensureValidImageUrl(mapImage);

  return (
    <AnimatePresence mode="wait">
      <CardContainer
        initial="hidden"
        animate="visible"
        exit="exit"
        variants={mapCardVariants}
        key={map.uuid}
        role="article"
        aria-label={`Map: ${map.displayName}`}
      >
        <MapImage
          src={mapImage}
          role="img"
          aria-label={`${map.displayName} map image`}
        />
        <CardContent>
          <MapName>{map.displayName}</MapName>
        </CardContent>
      </CardContainer>
    </AnimatePresence>
  );
};

export default React.memo(MapCard);
