export const getColor1 = function(input: string) {
  return "hsl(" + 12 * Math.floor(30 * cyrb53(input+"a")) + ',' +
    (35 + 10 * Math.floor(5 * cyrb53(input+"b"))) + '%,' +
    (25 + 10 * Math.floor(5 * cyrb53(input+"c"))) + '%)'
}

export const getColor2 = function(input: string) {
  return "hsl(" + 12 * Math.floor(30 * cyrb53(input+"a")) + ',' +
    (35 + 10 * Math.floor(5 * cyrb53(input+"b"))) + '%,' +
    '90%)'
}

//https://stackoverflow.com/questions/7616461/generate-a-hash-from-string-in-javascript?rq=1
const cyrb53 = function(str, seed = 0) {
  let h1 = 0xdeadbeef ^ seed, h2 = 0x41c6ce57 ^ seed;
  for (let i = 0, ch; i < str.length; i++) {
    ch = str.charCodeAt(i);
    h1 = Math.imul(h1 ^ ch, 2654435761);
    h2 = Math.imul(h2 ^ ch, 1597334677);
  }
  h1 = Math.imul(h1 ^ (h1>>>16), 2246822507) ^ Math.imul(h2 ^ (h2>>>13), 3266489909);
  h2 = Math.imul(h2 ^ (h2>>>16), 2246822507) ^ Math.imul(h1 ^ (h1>>>13), 3266489909);
  let hash = 4294967296 * (2097151 & h2) + (h1>>>0);
  return hash / Number.MAX_SAFE_INTEGER;
};
