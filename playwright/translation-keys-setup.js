import {test as setup} from '@playwright/test';

setup('enable translation keys', async ({ page }) => {
    await page.goto('/fixtures');
    await page.getByRole('link', {name: 'Toggle translation keys'}).click();
    await page.context().storageState({ path: './playwright/translation-keys-state.json' });
});
