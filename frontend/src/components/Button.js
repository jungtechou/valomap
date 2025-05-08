import React, { forwardRef } from "react";
import styled, { css } from "styled-components";
import { motion } from "framer-motion";
import PropTypes from "prop-types";

// Base button styles shared across all button variants
const baseButtonStyles = css`
  font-family: "Poppins", sans-serif;
  font-size: 0.9rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 2px;
  padding: 0.8rem 2rem;
  border-radius: 4px;
  position: relative;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;

  &:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  &:focus-visible {
    outline: 2px solid var(--valorant-white);
    outline-offset: 2px;
  }
`;

// Primary button
const PrimaryButton = styled(motion.button)`
  ${baseButtonStyles}
  background-color: var(--valorant-red);
  color: var(--valorant-white);
  border: 2px solid var(--valorant-red);

  &:hover:not(:disabled) {
    background-color: transparent;
    color: var(--valorant-red);
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(255, 70, 85, 0.3);
  }

  &:active:not(:disabled) {
    transform: translateY(0);
    box-shadow: none;
  }
`;

// Secondary button
const SecondaryButton = styled(motion.button)`
  ${baseButtonStyles}
  background-color: transparent;
  color: var(--valorant-white);
  border: 2px solid var(--valorant-white);

  &:hover:not(:disabled) {
    background-color: var(--valorant-white);
    color: var(--valorant-blue);
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(236, 232, 225, 0.2);
  }

  &:active:not(:disabled) {
    transform: translateY(0);
    box-shadow: none;
  }
`;

// Button content wrapper
const ButtonContent = styled.span`
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  z-index: 2;

  svg {
    font-size: 1.1em;
  }
`;

// Button animations
const buttonAnimations = {
  tap: { scale: 0.95 },
  hover: { scale: 1.05 },
  initial: { scale: 1 }
};

// Button component with ref forwarding
const Button = forwardRef(
  (
    {
      children,
      variant = "primary",
      onClick,
      disabled = false,
      type = "button",
      ...props
    },
    ref
  ) => {
    // Select button component based on variant
    const ButtonComponent = variant === "primary" ? PrimaryButton : SecondaryButton;

    return (
      <ButtonComponent
        ref={ref}
        onClick={onClick}
        disabled={disabled}
        type={type}
        whileTap={disabled ? {} : buttonAnimations.tap}
        whileHover={disabled ? {} : buttonAnimations.hover}
        initial={buttonAnimations.initial}
        transition={{ type: "spring", stiffness: 400, damping: 10 }}
        {...props}
      >
        <ButtonContent>{children}</ButtonContent>
      </ButtonComponent>
    );
  }
);

// Display name for debugging
Button.displayName = "Button";

// PropTypes for better development experience
Button.propTypes = {
  children: PropTypes.node.isRequired,
  variant: PropTypes.oneOf(["primary", "secondary"]),
  onClick: PropTypes.func,
  disabled: PropTypes.bool,
  type: PropTypes.oneOf(["button", "submit", "reset"]),
};

export default Button;
