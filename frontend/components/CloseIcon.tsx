import React from 'react';

export function CloseIcon({ size = 16, className = "", ...props }: React.SVGProps<SVGSVGElement> & { size?: number }) {
  return (
    <svg 
      width={size} 
      height={size} 
      viewBox="0 0 16 16" 
      fill="none" 
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      {...props}
    >
      <path
        d="M3.33398 3.33334L12.6673 12.6667M12.6673 3.33334L3.33398 12.6667"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="square"
      />
    </svg>
  );
}

