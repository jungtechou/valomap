import React from "react";
import styled from "styled-components";
import { FaGithub } from "react-icons/fa";

const HeaderContainer = styled.header`
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem 2rem;
  background-color: rgba(15, 25, 35, 0.95);
  border-bottom: 2px solid var(--valorant-red);

  @media (max-width: 768px) {
    padding: 1rem;
    flex-direction: column;
    gap: 1rem;
  }
`;

const Logo = styled.div`
  font-size: 1.5rem;
  font-weight: 700;
  letter-spacing: 1px;
  color: var(--valorant-white);
  display: flex;
  align-items: center;

  span {
    color: var(--valorant-red);
    font-weight: 700;
    padding-left: 5px;
  }
`;

const Nav = styled.nav`
  display: flex;
  gap: 1.5rem;
  align-items: center;
`;

const NavLink = styled.a`
  color: var(--valorant-white);
  font-weight: 500;
  font-size: 0.9rem;
  text-transform: uppercase;
  letter-spacing: 1px;
  position: relative;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  gap: 0.5rem;

  &:hover {
    color: var(--valorant-red);
  }

  &::after {
    content: "";
    position: absolute;
    bottom: -5px;
    left: 0;
    width: 0;
    height: 2px;
    background-color: var(--valorant-red);
    transition: width 0.3s ease;
  }

  &:hover::after {
    width: 100%;
  }
`;

const Header = () => {
  return (
    <HeaderContainer>
      <Logo>
        VALO<span>MAP</span>
      </Logo>
      <Nav>
        <NavLink href="https://github.com/jungtechou/valomap" target="_blank">
          <FaGithub /> Source
        </NavLink>
      </Nav>
    </HeaderContainer>
  );
};

export default Header;
