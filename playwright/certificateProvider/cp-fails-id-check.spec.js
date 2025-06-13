import {expect, test} from '@playwright/test';
import {screenshot} from '../../../test/playwright/e2e.js'
import {extractTextFromMainAndSave} from "../../../test/playwright/textExtractor.js";

test('certificate provider fails ID check', async ({page}) => {
    await page.goto('/fixtures');
    await page.getByRole('link', {name: 'Certificate provider'}).click();
    await page.getByRole('radio', {name: 'Confirm your details'}).check();
    await page.getByRole('button', {name: /Start|Start now/}).click();
    await page.getByRole('tab', {name: 'I’m a certificate provider'}).click();
    await page.getByRole('link', {name: 'Go to task list'}).click();
    await page.getByRole('link', {name: 'Confirm your identity'}).click();
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('radio', {name: 'Failed identity check (T)'}).check();
    await page.getByRole('button', {name: 'Continue'}).click();

    await expect(page.getByLabel('Important').getByRole('paragraph')).toContainText('You were not able to confirm your identity using GOV.UK One Login.');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Return to task list'}).click();
    await page.getByRole('link', {name: 'Provide your certificate'}).click();
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('link', {name: 'Continue'}).click();
    await page.getByRole('checkbox', {name: 'I, Charlie Cooper, agree'}).check();
    await page.getByRole('button', {name: 'Submit signature'}).click();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
});
