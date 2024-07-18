describe('People to notify', () => {
    it('can add people to notify', () => {
        cy.visit('/fixtures?redirect=/do-you-want-to-notify-people&progress=chooseYourAttorneys');

        cy.checkA11yApp();

        cy.get('input[name="yes-no"]').check('yes', { force: true }, { force: true });
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/choose-people-to-notify');

        cy.checkA11yApp();

        cy.get('#f-first-names').type("Brian")
        cy.get('#f-last-name').type("Gooding")

        cy.contains('button', 'Save and continue').click();

        cy.contains('label', 'Enter a new address').click();
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/choose-people-to-notify-address');
        cy.checkA11yApp();

        cy.get('#f-lookup-postcode').type("B14 7ED")
        cy.contains('button', 'Find address').click();

        cy.url().should('contain', '/choose-people-to-notify-address');
        cy.checkA11yApp();

        cy.contains('a', "I can’t find their address in the list").click();

        cy.url().should('contain', '/choose-people-to-notify-address');
        cy.checkA11yApp();

        cy.get('#f-address-line-1').type("4 RICHMOND PLACE");
        cy.get('#f-address-town').type("BIRMINGHAM");
        cy.get('#f-address-postcode').type("B14 7ED");

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/choose-people-to-notify-summary');

        cy.checkA11yApp();

        cy.contains('People to notify about your LPA');

        cy.contains('.govuk-summary-card', 'Brian Gooding').within(() => {
            cy.contains('Brian Gooding');
            cy.contains("4 RICHMOND PLACE");
            cy.contains("BIRMINGHAM");
            cy.contains("B14 7ED");
        });

        cy.get('input[name="yes-no"]').check('yes', { force: true })
        cy.contains('button', 'Save and continue').click();

        cy.visitLpa('/task-list')

        cy.contains('a', 'People to notify').parent().parent().contains('1 added')
    });

    it('can modify a person to notifys details', () => {
        cy.visit('/fixtures?redirect=/choose-people-to-notify-summary&progress=peopleToNotifyAboutYourLpa');

        cy.checkA11yApp();

        cy.contains('.govuk-summary-list__row', 'Jordan Jefferson').contains('a', 'Change').click();

        cy.url().should('contain', '/choose-people-to-notify');

        cy.checkA11yApp();

        cy.get('#f-first-names').clear().type('Changed')
        cy.get('#f-last-name').clear().type('Altered')

        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/choose-people-to-notify-summary');

        cy.contains('.govuk-summary-card', 'Changed Altered');
        cy.contains('.govuk-summary-list__row', '4 RICHMOND PLACE').contains('a', 'Change').click();

        cy.url().should('contain', '/choose-people-to-notify-address');

        cy.checkA11yApp();

        cy.get('#f-address-line-1').clear().type('1 New Road');
        cy.get('#f-address-line-2').clear().type('Changeville');
        cy.get('#f-address-line-3').clear().type('Newington');
        cy.get('#f-address-town').clear().type('Newshire');
        cy.get('#f-address-postcode').clear().type('A12 3BC');

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/choose-people-to-notify-summary');

        cy.contains('.govuk-summary-card', 'Changed Altered').within(() => {
            cy.contains('1 New Road');
            cy.contains('Changeville');
            cy.contains('Newington');
            cy.contains('Newshire');
            cy.contains('A12 3BC');
        });
    });

    it('can remove a person to notify', () => {
        cy.visit('/fixtures?redirect=/choose-people-to-notify-summary&progress=peopleToNotifyAboutYourLpa');

        cy.checkA11yApp();

        cy.contains('.govuk-summary-card', 'Danni Davies').contains('Remove person to notify').click();

        cy.url().should('contain', '/remove-person-to-notify');

        cy.checkA11yApp();

        cy.get('input[name="yes-no"]').check('yes', { force: true })
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/choose-people-to-notify-summary');

        cy.contains('.govuk-summary-card', 'Jordan Jefferson').contains('Remove person to notify').click();

        cy.url().should('contain', '/remove-person-to-notify');

        cy.get('input[name="yes-no"]').check('yes', { force: true })
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/do-you-want-to-notify-people');
    });

    it('hides action links when LPA has been signed', () => {
        cy.visit('/fixtures?redirect=/choose-people-to-notify-summary&progress=signTheLpa');

        cy.checkA11yApp();

        cy.contains('Jordan Jefferson').parent().contains('a', 'Change').should('not.exist');
    });

    it('limits people to notify to 5', () => {
        cy.visit('/fixtures?redirect=/choose-people-to-notify-summary&progress=peopleToNotifyAboutYourLpa&peopleToNotify=max');

        cy.checkA11yApp();

        cy.contains('Do you want to add another person to notify?').should('not.exist');

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/task-list');
    });

    it('errors when unselected', () => {
        cy.visit('/fixtures?redirect=/do-you-want-to-notify-people&progress=chooseYourAttorneys');
        cy.contains('button', 'Save and continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select yes to notify someone about your LPA');
        });

        cy.contains('.govuk-fieldset .govuk-error-message', 'Select yes to notify someone about your LPA');
    });

    it('errors when people to notify details empty', () => {
        cy.visit('/fixtures?redirect=/choose-people-to-notify&progress=chooseYourAttorneys');
        cy.contains('button', 'Save and continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter first names');
            cy.contains('Enter last name');
        });

        cy.contains('[for=f-first-names] + div + .govuk-error-message', 'Enter first names');
        cy.contains('[for=f-last-name] + .govuk-error-message', 'Enter last name');
    });

    it('errors when people to notify details invalid', () => {
        cy.visit('/fixtures?redirect=/choose-people-to-notify&progress=chooseYourAttorneys');

        cy.get('#f-first-names').invoke('val', 'a'.repeat(54));
        cy.get('#f-last-name').invoke('val', 'b'.repeat(62));
        cy.contains('button', 'Save and continue').click();

        cy.contains('[for=f-first-names] + div + .govuk-error-message', 'First names must be 53 characters or less');
        cy.contains('[for=f-last-name] + .govuk-error-message', 'Last name must be 61 characters or less');
    });

    it('errors when another not selected', () => {
        cy.visit('/fixtures?redirect=/choose-people-to-notify-summary&progress=peopleToNotifyAboutYourLpa');

        cy.checkA11yApp();

        cy.contains('button', 'Save and continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select yes to add another person to notify');
        });

        cy.contains('.govuk-fieldset .govuk-error-message', 'Select yes to add another person to notify');
    });

    it('warns when name shared with other actor', () => {
        cy.visit('/fixtures?redirect=/choose-people-to-notify&progress=chooseYourAttorneys');

        cy.get('#f-first-names').type('Sam');
        cy.get('#f-last-name').type('Smith');
        cy.contains('button', 'Save and continue').click();
        cy.url().should('contain', '/choose-people-to-notify');

        cy.contains('The donor’s name is also Sam Smith.');

        cy.contains('button', 'Save and continue').click();
        cy.url().should('contain', '/choose-people-to-notify-address');
    });
});
