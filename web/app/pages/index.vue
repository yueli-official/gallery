<script setup lang="ts">
const site = useGallerySiteSettings();
useSeoMeta({ title: () => site.value?.title || "月离图库", description: () => site.value?.description || "收集、分类与发现喜欢的图片。" });
</script>
<template><GalleryCatalog /></template>
