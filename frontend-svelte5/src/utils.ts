export const emailValidationPattern = "[^@\\s]+@[^@\\s]+\\.[^@\\s]+";

export const getColor1 = function (input: string) {
  return (
    "hsl(" +
    12 * Math.floor(30 * cyrb53(input + "a")) +
    "," +
    (35 + 10 * Math.floor(5 * cyrb53(input + "b"))) +
    "%," +
    (25 + 10 * Math.floor(5 * cyrb53(input + "c"))) +
    "%)"
  );
};

export const getColor2 = function (input: string) {
  return (
    "hsl(" +
    12 * Math.floor(30 * cyrb53(input + "a")) +
    "," +
    (35 + 10 * Math.floor(5 * cyrb53(input + "b"))) +
    "%," +
    "90%)"
  );
};

//https://stackoverflow.com/questions/7616461/generate-a-hash-from-string-in-javascript?rq=1
const cyrb53 = function (str: string, seed = 0) {
  let h1 = 0xdeadbeef ^ seed,
    h2 = 0x41c6ce57 ^ seed;
  for (let i = 0, ch; i < str.length; i++) {
    ch = str.charCodeAt(i);
    h1 = Math.imul(h1 ^ ch, 2654435761);
    h2 = Math.imul(h2 ^ ch, 1597334677);
  }
  h1 =
    Math.imul(h1 ^ (h1 >>> 16), 2246822507) ^
    Math.imul(h2 ^ (h2 >>> 13), 3266489909);
  h2 =
    Math.imul(h2 ^ (h2 >>> 16), 2246822507) ^
    Math.imul(h1 ^ (h1 >>> 13), 3266489909);
  let hash = 4294967296 * (2097151 & h2) + (h1 >>> 0);
  return hash / Number.MAX_SAFE_INTEGER;
};

/*export const htmlLegendPlugin = {
  id: "htmlLegend",
  afterUpdate(chart, args, options) {
    const ul = getOrCreateLegendList(chart, options.containerID);

    // Remove old legend items
    while (ul.firstChild) {
      ul.firstChild.remove();
    }

    // Reuse the built-in legendItems generator
    const items = chart.options.plugins.legend.labels.generateLabels(chart);

    items.forEach((item) => {
      const li = document.createElement("li");
      li.style.alignItems = "center";
      li.style.cursor = "pointer";
      li.style.display = "flex";
      li.style.flexDirection = "row";
      li.style.marginLeft = "10px";
      li.style.float = "left";

      li.onclick = () => {
        chart.setDatasetVisibility(
          item.datasetIndex,
          !chart.isDatasetVisible(item.datasetIndex)
        );
        chart.update();
      };

      // Color box
      const boxSpan = document.createElement("span");
      boxSpan.style.background = item.fillStyle;
      boxSpan.style.borderColor = item.strokeStyle;
      boxSpan.style.borderWidth = item.lineWidth + "px";
      boxSpan.style.display = "inline-block";
      boxSpan.style.height = "20px";
      boxSpan.style.marginRight = "10px";
      boxSpan.style.width = "20px";

      // Text
      const textContainer = document.createElement("p");
      textContainer.style.color = item.fontColor;
      textContainer.style.margin = "0";
      textContainer.style.padding = "0";
      textContainer.style.textDecoration = item.hidden ? "line-through" : "";

      let start = item.text.indexOf(";");
      let label = item.text.substring(0, start);
      const text = document.createTextNode(label);
      textContainer.appendChild(text);

      li.appendChild(boxSpan);
      li.appendChild(textContainer);
      ul.appendChild(li);
    });
  },
};

const getOrCreateLegendList = (chart, id) => {
  const legendContainer = document.getElementById(id);
  let listContainer = legendContainer.querySelector("ul");

  if (!listContainer) {
    listContainer = document.createElement("ul");
    listContainer.style.margin = "0";
    listContainer.style.padding = "0";

    legendContainer.appendChild(listContainer);
  }

  return listContainer;
};*/

//https://github.com/terkelg/skaler, MIT license
interface ScalerOptions {
  scale?: number;
  width?: number;
  height?: number;
  name?: string;
  type?: string;
  quality?: number
}

export default function skaler(file: File, options: ScalerOptions = {}): Promise<File> {
  const {
    scale,
    width,
    height,
    name = file.name,
    type = file.type,
    quality,
  } = options;

  return new Promise((res, rej) => {
    const reader = new FileReader();
    reader.readAsDataURL(file);

    reader.onload = (e: ProgressEvent<FileReader>) => {
      const img = new Image();

      img.onload = () => {
        const el = document.createElement('canvas');
        const dir = (width && width < img.width) || (height && height < img.height) ? 'min' : 'max';
        const stretch = width && height;

        const ratio = scale ? scale : Math[dir](
            (width ? width / img.width : 1),
            (height ? height / img.height : 1)
        );

        const w = el.width = stretch ? (width ?? img.width) : img.width * ratio;
        const h = el.height = stretch ? (height ?? img.height) : img.height * ratio;

        const ctx = el.getContext('2d');
        if (!ctx) {
          rej(new Error('Failed to get canvas context'));
          return;
        }

        ctx.imageSmoothingEnabled = true;
        ctx.imageSmoothingQuality = "high";
        ctx.drawImage(img, 0, 0, w, h);

        el.toBlob((blob) => {
          if (!blob) {
            rej(new Error('Failed to create blob'));
            return;
          }
          res(new File([blob], name, { type, lastModified: Date.now() }));
        }, "image/jpeg", quality);
      };
      //https://developer.mozilla.org/en-US/docs/Web/API/HTMLCanvasElement/toBlob

      img.onerror = () => rej(new Error('Failed to load image'));

      const result = e.target?.result;
      if (typeof result !== 'string') {
        rej(new Error('Failed to read file'));
        return;
      }
      img.src = result;
    };

    reader.onerror = () => rej(new Error('Failed to read file'));
  });
}

export function debounce<T extends (...args: any[]) => any>(
    func: T,
    wait: number
): (...args: Parameters<T>) => void {
  let timeout: ReturnType<typeof setTimeout>;

  return (...args: Parameters<T>) => {
    clearTimeout(timeout);
    timeout = setTimeout(() => func(...args), wait);
  };
}