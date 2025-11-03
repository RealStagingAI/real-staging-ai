'use client';

import { useState, useRef, MouseEvent, TouchEvent } from 'react';
import Image from 'next/image';
import { cn } from '@/lib/utils';

interface BeforeAfterSliderProps {
  beforeSrc: string;
  afterSrc: string;
  beforeAlt?: string;
  afterAlt?: string;
  className?: string;
  initialPosition?: number;
}

export default function BeforeAfterSlider({
  beforeSrc,
  afterSrc,
  beforeAlt = 'Before',
  afterAlt = 'After',
  className,
  initialPosition = 50,
}: BeforeAfterSliderProps) {
  const [position, setPosition] = useState(initialPosition);
  const [isDragging, setIsDragging] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  const updatePosition = (clientX: number) => {
    if (!containerRef.current) return;
    
    const rect = containerRef.current.getBoundingClientRect();
    const x = clientX - rect.left;
    const percentage = Math.max(0, Math.min(100, (x / rect.width) * 100));
    setPosition(percentage);
  };

  const handleMouseDown = (e: MouseEvent) => {
    setIsDragging(true);
    updatePosition(e.clientX);
  };

  const handleMouseMove = (e: MouseEvent) => {
    if (!isDragging) return;
    updatePosition(e.clientX);
  };

  const handleMouseUp = () => {
    setIsDragging(false);
  };

  const handleTouchStart = (e: TouchEvent) => {
    setIsDragging(true);
    updatePosition(e.touches[0].clientX);
  };

  const handleTouchMove = (e: TouchEvent) => {
    if (!isDragging) return;
    updatePosition(e.touches[0].clientX);
  };

  const handleTouchEnd = () => {
    setIsDragging(false);
  };

  return (
    <div
      ref={containerRef}
      className={cn(
        'relative overflow-hidden rounded-2xl shadow-2xl cursor-ew-resize select-none',
        'aspect-[4/3] sm:aspect-[16/9] w-full',
        isDragging && 'cursor-grabbing',
        className
      )}
      onMouseDown={handleMouseDown}
      onMouseMove={handleMouseMove}
      onMouseUp={handleMouseUp}
      onMouseLeave={handleMouseUp}
      onTouchStart={handleTouchStart}
      onTouchMove={handleTouchMove}
      onTouchEnd={handleTouchEnd}
    >
      {/* Before Image */}
      <div className="absolute inset-0">
        <div className="relative w-full h-full">
          <Image
            src={beforeSrc}
            alt={beforeAlt}
            fill
            className="object-cover"
            draggable={false}
            sizes="(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw"
          />
        </div>
        <div className="absolute top-4 left-4 bg-black/60 text-white px-3 py-1.5 rounded-lg text-sm font-medium backdrop-blur-sm z-10">
          Before
        </div>
      </div>

      {/* After Image (clipped) */}
      <div
        className="absolute inset-0"
        style={{
          clipPath: `inset(0 ${100 - position}% 0 0)`,
        }}
      >
        <div className="relative w-full h-full">
          <Image
            src={afterSrc}
            alt={afterAlt}
            fill
            className="object-cover"
            draggable={false}
            sizes="(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw"
          />
        </div>
        <div className="absolute top-4 right-4 bg-black/60 text-white px-3 py-1.5 rounded-lg text-sm font-medium backdrop-blur-sm z-10">
          After
        </div>
      </div>

      {/* Subtle overlay to help blend minor misalignments */}
      <div 
        className="absolute inset-0 pointer-events-none"
        style={{
          background: `linear-gradient(to right, 
            transparent 0%, 
            transparent ${Math.max(0, position - 2)}%, 
            rgba(255,255,255,0.03) ${Math.max(0, position - 2)}%, 
            rgba(255,255,255,0.03) ${Math.min(100, position + 2)}%, 
            transparent ${Math.min(100, position + 2)}%, 
            transparent 100%
          )`,
        }}
      />

      {/* Slider Line */}
      <div
        className="absolute top-0 bottom-0 w-0.5 bg-white shadow-lg pointer-events-none"
        style={{ left: `${position}%` }}
      >
        {/* Slider Handle */}
        <div className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-12 h-12 bg-white rounded-full shadow-lg border-2 border-gray-200 flex items-center justify-center pointer-events-auto">
          <svg
            className="w-6 h-6 text-gray-600"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M8 9l4-4 4 4m0 6l-4 4-4-4"
            />
          </svg>
        </div>
      </div>

      {/* Instruction Text */}
      <div className="absolute bottom-4 left-1/2 transform -translate-x-1/2 bg-black/60 text-white px-4 py-2 rounded-lg text-sm backdrop-blur-sm pointer-events-none">
        Slide to compare
      </div>
    </div>
  );
}
