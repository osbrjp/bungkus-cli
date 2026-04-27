<script setup lang="ts">
/**
 * Layout: Default
 *
 * @file ./app/layouts/default.vue
 * @description Default page layout. Renders the site header, slotted page content, and site footer.
 * @module Layout
 */
</script>

<template>
  <div>
    <LayoutHeader />
    <main>
      <slot />
    </main>
    <LayoutFooter />
  </div>
</template>

