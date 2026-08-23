<script lang="ts">
export type BadgeTone =
  | 'neutral'
  | 'primary'
  | 'info'
  | 'success'
  | 'warning'
  | 'destructive'
  | 'outline'
</script>

<script setup lang="ts">
import { cva, type VariantProps } from 'class-variance-authority'

const badgeVariants = cva(
  'inline-flex items-center gap-1 rounded-full border px-2.5 py-0.5 text-xs font-medium tracking-wide transition-colors',
  {
    variants: {
      tone: {
        neutral: 'border-border bg-secondary text-secondary-foreground',
        primary: 'border-transparent bg-primary text-primary-foreground',
        info: 'border-primary/30 bg-primary/10 text-primary',
        success: 'border-transparent bg-success/15 text-success',
        warning: 'border-transparent bg-warning/15 text-warning',
        destructive: 'border-transparent bg-destructive/10 text-destructive',
        outline: 'border-border bg-transparent text-foreground',
      },
      uppercase: {
        true: 'uppercase',
        false: '',
      },
    },
    defaultVariants: {
      tone: 'neutral',
      uppercase: false,
    },
  },
)

interface Props {
  tone?: BadgeTone
  uppercase?: boolean
}

const props = withDefaults(defineProps<Props>(), { tone: undefined, uppercase: undefined })

type VariantTone = NonNullable<VariantProps<typeof badgeVariants>['tone']>

const resolved = (): { tone?: VariantTone; uppercase?: boolean } => ({
  tone: props.tone as VariantTone | undefined,
  uppercase: props.uppercase,
})
</script>

<template>
  <span :class="badgeVariants(resolved())">
    <slot />
  </span>
</template>
