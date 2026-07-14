<script setup lang="ts">
definePageMeta({ layout: "manage", middleware: "auth" });
await navigateTo("/manage/reviews", { redirectCode: 302 });
</script>
