<style>
  #dots #dot1 {
    animation: load 1s infinite;
  }

  #dots #dot2 {
    animation: load 1s infinite;
    animation-delay: 0.2s;
  }

  #dots #dot3 {
    animation: load 1s infinite;
    animation-delay: 0.4s;
  }

  @keyframes load {
    0% {
      opacity: 0;
    }
    50% {
      opacity: 1;
    }
    100% {
      opacity: 0;
    }
  }
</style>

<span id="dots">
  <span id="dot1">.</span><span id="dot2">.</span><span id="dot3">.</span>
</span>
