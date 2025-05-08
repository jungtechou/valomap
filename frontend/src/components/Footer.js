import React from "react";
import styled from "styled-components";

const FooterContainer = styled.footer`
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 1.5rem;
  background-color: rgba(15, 25, 35, 0.95);
  border-top: 2px solid var(--valorant-red);
  font-size: 0.85rem;
`;

const FooterText = styled.p`
  color: var(--valorant-light-gray);
  text-align: center;
  line-height: 1.5;

  a {
    color: var(--valorant-red);
    transition: color 0.2s ease;

    &:hover {
      color: var(--valorant-white);
    }
  }
`;

const Footer = () => {
  const currentYear = new Date().getFullYear();

  return (
    <FooterContainer>
      <FooterText>
        © {currentYear} ValoMap. Not affiliated with Riot Games.
        <br />
        VALORANT and Riot Games are trademarks or registered trademarks of Riot
        Games, Inc.
      </FooterText>
    </FooterContainer>
  );
};

export default Footer;
