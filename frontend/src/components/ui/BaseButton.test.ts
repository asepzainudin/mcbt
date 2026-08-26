import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import BaseButton from './BaseButton.vue'

describe('BaseButton', () => {
  it('merender teks dari slot default', () => {
    const wrapper = mount(BaseButton, { slots: { default: 'Simpan Data' } })
    expect(wrapper.text()).toContain('Simpan Data')
  })

  it('memakai class variant primary secara default', () => {
    const wrapper = mount(BaseButton)
    expect(wrapper.find('button').classes()).toContain('bg-primary')
  })

  it('memakai class variant outline saat diminta', () => {
    const wrapper = mount(BaseButton, { props: { variant: 'outline' } })
    expect(wrapper.find('button').classes()).toContain('border-input')
  })

  it('menonaktifkan tombol saat loading', () => {
    const wrapper = mount(BaseButton, { props: { loading: true } })
    expect(wrapper.find('button').attributes('disabled')).toBeDefined()
    expect(wrapper.find('.animate-spin').exists()).toBe(true)
  })

  it('type default adalah button (bukan submit)', () => {
    const wrapper = mount(BaseButton)
    expect(wrapper.find('button').attributes('type')).toBe('button')
  })
})
