import {expect, test} from '@playwright/test';
import {randomShareCode, screenshot, TestEmail, TestMobile} from '../../../test/playwright/e2e.js'
import {extractTextFromMainAndSave} from "../../../test/playwright/textExtractor.js";

test('voucher completes their journey', async ({page}) => {
    const shareCode = randomShareCode()

    await page.goto(`/fixtures/voucher?redirect=&progress=&withShareCode=${shareCode}&email=${TestEmail}&donorMobile=${TestMobile}`);

    await expect(page.locator('h1')).toContainText('What is vouching?');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: /Start|Start now/}).click();
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('heading', {name: 'Vouch for someone'}).click();

    await expect(page.locator('h1')).toContainText('Vouch for someone');
    await page.getByRole('textbox', {name: 'Enter code'}).fill(shareCode);
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Save and continue'}).click();

    await expect(page.locator('#main-content')).toContainText('Vouch for someone’s identity Your task list');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Confirm your name'}).click();

    await expect(page.locator('h1')).toContainText('Confirm your name');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('link', {name: 'Change   last name'}).click();

    await expect(page.locator('h1')).toContainText('Your name');
    await page.getByRole('textbox', {name: 'Last name'}).fill('Smith');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Save and continue'}).click();
    await page.getByRole('button', {name: 'Continue'}).click();

    await expect(page.locator('#main-content')).toContainText('Confirm that you are allowed to vouch');
    await page.getByRole('radio', {name: 'Yes'}).check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('link', {name: 'Verify Sam Smith’s identity'}).click();

    await expect(page.locator('h1')).toContainText('Verify Sam Smith’s identity');
    await page.getByRole('radio', {name: 'Yes'}).check();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('link', {name: 'Confirm your identity'}).click();

    await expect(page.locator('h1')).toContainText('Confirm your identity');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('radio', {name: 'Custom'}).check();
    await page.getByRole('textbox', {name: 'First names'}).fill('Vivian');
    await page.getByRole('textbox', {name: 'Last name'}).fill('Smith');
    await page.getByRole('textbox', {name: 'Day'}).fill('23');
    await page.getByRole('textbox', {name: 'Month'}).fill('01');
    await page.getByRole('textbox', {name: 'Year'}).fill('1992');
    await page.getByRole('textbox', {name: 'Street name'}).fill('1 Darnick Raoad');
    await page.getByRole('textbox', {name: 'Town or city'}).fill('Birmingham');
    await page.getByRole('textbox', {name: 'Postcode'}).fill('B73 6PE');
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('radio', {name: 'Yes'}).check();
    await page.getByRole('button', {name: 'Continue'}).click();
    await page.getByRole('link', {name: 'Sign the declaration'}).click();

    await expect(page.locator('h1')).toContainText('Your declaration');
    await page.getByRole('checkbox', {name: 'To the best of my knowledge,'}).check();
    await page.getByText('How ticking the box acts as').click();
    await screenshot(page)
    await extractTextFromMainAndSave(page)
    await page.getByRole('button', {name: 'Submit my signature'}).click();

    await expect(page.locator('#main-content')).toContainText('You have vouched for Sam Smith’s identity.');
    await screenshot(page)
    await extractTextFromMainAndSave(page)
});
