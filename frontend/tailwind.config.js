module.exports = {
  future: {
    // removeDeprecatedGapUtilities: true,
    purgeLayersByDefault: true,
  },
  purge: ["./src/**/*.html", "./src/**/*.svelte"],
  theme: {
    fontFamily: {
      openSans: ['"Open Sans"', "sans-serif"],
      raleway: ["Raleway", "serif"],
    },
    extend: {
      colors: {
        red: {
          50: "#FAF5F3",
          100: "#F5EBE7",
          200: "#E6CEC4",
          300: "#D6B1A0",
          400: "#B87658",
          500: "#993B11",
          600: "#8A350F",
          700: "#5C230A",
          800: "#451B08",
          900: "#2E1205",
        },
        primary: {
          50: "#F6F9F7",
          100: "#ECF3EF",
          200: "#D0E2D7",
          300: "#B4D0BF",
          400: "#7BAD8E",
          500: "#438A5E",
          600: "#3C7C55",
          700: "#285338",
          800: "#1E3E2A",
          900: "#14291C",
        },
        secondary: {
          50: "#FCFCF7",
          100: "#F8FAF0",
          200: "#EEF2D8",
          300: "#E3E9C1",
          400: "#CFD993",
          500: "#BAC964",
          600: "#A7B55A",
          700: "#70793C",
          800: "#545A2D",
          900: "#383C1E",
        },
        light: {
          50: "#FFFFFE",
          100: "#FEFFFC",
          200: "#FDFEF8",
          300: "#FCFDF3",
          400: "#F9FCEA",
          500: "#F7FBE1",
          600: "#DEE2CB",
          700: "#949787",
          800: "#6F7165",
          900: "#4A4B44",
        },
        blue: {
          50: "#F6F8F9",
          100: "#ECF1F3",
          200: "#D0DBE2",
          300: "#B4C5D0",
          400: "#7B9AAD",
          500: "#436F8A",
          600: "#3C647C",
          700: "#284353",
          800: "#1E323E",
          900: "#142129",
        },
      },
    },
  },
  variants: {},
  plugins: [],
};
