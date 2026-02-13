/**
 * Ripple Button Component
 * Button with material design ripple effect on click
 */
import { forwardRef } from 'react';
import clsx from 'clsx';

/**
 * RippleButton - A button component with ripple click animation
 * @param {object} props - Component props
 * @param {React.ReactNode} props.children - Button content
 * @param {Function} props.onClick - Click handler
 * @param {string} props.className - Additional CSS classes
 * @param {string} props.rippleColor - Ripple effect color (default: rgba(255, 255, 255, 0.5))
 * @param {boolean} props.disabled - Disabled state
 */
const RippleButton = forwardRef(({ 
  children, 
  onClick, 
  className, 
  rippleColor = 'rgba(255, 255, 255, 0.5)',
  disabled = false,
  ...rest 
}, ref) => {
  const handleClick = (e) => {
    if (disabled) return;

    const button = e.currentTarget;
    const ripple = document.createElement('span');
    const rect = button.getBoundingClientRect();
    
    // Calculate ripple size and position
    const size = Math.max(rect.width, rect.height);
    const x = e.clientX - rect.left - size / 2;
    const y = e.clientY - rect.top - size / 2;
    
    // Style the ripple element
    ripple.style.cssText = `
      position: absolute;
      width: ${size}px;
      height: ${size}px;
      border-radius: 50%;
      background: ${rippleColor};
      transform: scale(0);
      animation: ripple 600ms ease-out;
      left: ${x}px;
      top: ${y}px;
      pointer-events: none;
    `;
    
    // Add ripple to button
    button.appendChild(ripple);
    
    // Remove ripple after animation
    setTimeout(() => {
      ripple.remove();
    }, 600);
    
    // Call original onClick handler
    if (onClick) {
      onClick(e);
    }
  };
  
  return (
    <button 
      ref={ref}
      className={clsx('relative overflow-hidden ripple-button', className)}
      onClick={handleClick}
      disabled={disabled}
      {...rest}
    >
      {children}
    </button>
  );
});

RippleButton.displayName = 'RippleButton';

export default RippleButton;
