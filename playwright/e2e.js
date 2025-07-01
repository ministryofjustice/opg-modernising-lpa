import AxeBuilder from '@axe-core/playwright';
import { expect } from '@playwright/test';

export const
    TestEmail = 'simulate-delivered@notifications.service.gov.uk',
    TestEmail2 = 'simulate-delivered-2@notifications.service.gov.uk',
    TestMobile = '07700900000'

export function randomAccessCode() {
    const characters = 'abcdefghijklmnpqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789'
    let result = [];

    for (let i = 0; i < 12; i++) {
        result.push(characters.charAt(Math.floor(Math.random() * characters.length)));
    }

    return result.join('');
}

export async function checkA11y(page) {
    const accessibilityScanResults = await new AxeBuilder({ page })
        .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
        .exclude('.govuk-phase-banner')
        .analyze();

    expect(accessibilityScanResults.violations).toEqual([]);
}

export async function visitLpa(path, page) {
    console.log(page.url())
    await page.goto(page.url().split('/').slice(0, 5).join('/') + path);
}

export function sanitisedPath(page) {
    const parts = page.url().split('/')
    return parts[parts.length - 1].replace(/\//g, ' ')
}

export async function screenshot(page) {
    await page.screenshot({ path: `playwright/screenshots/${Date.now()} ${sanitisedPath(page)}.png`, fullPage: true });
}
