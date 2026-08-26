import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import BaseBadge from './BaseBadge.vue'

describe('BaseBadge', () => {
  it('merender teks dari slot', () => {
    const wrapper = mount(BaseBadge, { slots: { default: 'LULUS' } })
    expect(wrapper.text()).toContain('LULUS')
  })

  it('merender elemen span', () => {
    const wrapper = mount(BaseBadge)
    expect(wrapper.find('span').exists()).toBe(true)
  })
})
