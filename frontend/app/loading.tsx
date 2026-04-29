"use client";

import { motion } from "framer-motion";

export default function Loading() {
  return (
    <div className="min-h-screen w-full bg-canvas flex flex-col items-center justify-center gap-6">
      <div className="relative">
        {/* Outer glowing ring */}
        <motion.div
          animate={{
            scale: [1, 1.2, 1],
            opacity: [0.3, 0.6, 0.3],
          }}
          transition={{
            duration: 2,
            repeat: Infinity,
            ease: "easeInOut",
          }}
          className="absolute inset-0 rounded-full bg-brand/20 blur-xl"
        />
        
        {/* Main loader */}
        <div className="relative w-16 h-16">
          <motion.div
            animate={{ rotate: 360 }}
            transition={{
              duration: 1.5,
              repeat: Infinity,
              ease: "linear",
            }}
            className="w-full h-full rounded-full border-t-2 border-r-2 border-brand"
          />
          
          {/* Inner pulse */}
          <motion.div
            animate={{
              scale: [0.8, 1.1, 0.8],
            }}
            transition={{
              duration: 1.5,
              repeat: Infinity,
              ease: "easeInOut",
            }}
            className="absolute inset-4 rounded-full bg-brand"
          />
        </div>
      </div>
      
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.5 }}
        className="text-zinc-500 font-medium tracking-widest text-xs uppercase"
      >
        Loading Experience
      </motion.div>
    </div>
  );
}
