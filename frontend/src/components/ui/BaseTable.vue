<template>
  <div
    class="w-full overflow-x-auto rounded-xl border border-border bg-card text-card-foreground shadow-sm"
  >
    <table class="w-full caption-bottom text-sm">
      <thead>
        <tr class="border-b border-border bg-muted/50">
          <slot name="head" />
        </tr>
      </thead>
      <tbody>
        <slot />
      </tbody>
    </table>
  </div>
</template>
