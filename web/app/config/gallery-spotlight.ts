export interface GallerySpotlightCharacter {
  id: string;
  name: string;
  series: string;
  faction: string;
  image: string;
  thumb: string;
  accent: string;
  wash: string;
}

// Local design-review spotlight. These are demo character renders, not yet
// gallery records; production should move them behind the Asset pipeline.
export const gallerySpotlight: GallerySpotlightCharacter[] = [
  {
    id: "shiroko",
    name: "砂狼白子",
    series: "Blue Archive",
    faction: "阿拜多斯",
    image: "/spotlight/shiroko.webp",
    thumb: "/spotlight/shiroko-thumb.webp",
    accent: "#7fd8ff",
    wash: "radial-gradient(circle at 50% 34%, #17394a 0%, #090d12 68%)",
  },
  {
    id: "hoshino",
    name: "小鸟游星野",
    series: "Blue Archive",
    faction: "阿拜多斯",
    image: "/spotlight/hoshino.webp",
    thumb: "/spotlight/hoshino-thumb.webp",
    accent: "#ffd277",
    wash: "radial-gradient(circle at 50% 34%, #3d2a12 0%, #0d0b08 68%)",
  },
  {
    id: "aru",
    name: "陆八魔阿露",
    series: "Blue Archive",
    faction: "格黑娜",
    image: "/spotlight/aru.webp",
    thumb: "/spotlight/aru-thumb.webp",
    accent: "#ff7e95",
    wash: "radial-gradient(circle at 50% 34%, #3a1420 0%, #0e080a 68%)",
  },
  {
    id: "yuuka",
    name: "早濑优香",
    series: "Blue Archive",
    faction: "千年",
    image: "/spotlight/yuuka.webp",
    thumb: "/spotlight/yuuka-thumb.webp",
    accent: "#8dffd9",
    wash: "radial-gradient(circle at 50% 34%, #0e3a2e 0%, #060d0b 68%)",
  },
  {
    id: "mika",
    name: "圣园弥香",
    series: "Blue Archive",
    faction: "崔尼蒂",
    image: "/spotlight/mika.webp",
    thumb: "/spotlight/mika-thumb.webp",
    accent: "#ffb3e6",
    wash: "radial-gradient(circle at 50% 34%, #3a1a30 0%, #0e080d 68%)",
  },
];
