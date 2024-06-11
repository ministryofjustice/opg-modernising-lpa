describe('Check the LPA', () => {
  it('cannot change when personal welfare LPA can be used', () => {
    cy.visit('/fixtures?redirect=/check-your-lpa&progress=peopleToNotifyAboutYourLpa&lpa-type=personal-welfare');

    cy.contains('.govuk-summary-list__row', 'When your attorneys can use your LPA')
      .contains('Only when I do not have mental capacity')
      .contains('Change').should('not.exist');
  });

  it("can submit the completed LPA", () => {
    cy.visit('/fixtures?redirect=/check-your-lpa&progress=peopleToNotifyAboutYourLpa&certificateProviderEmail=test@example.com');

    cy.contains('h1', "Check your LPA")

    cy.checkA11yApp();

    cy.contains('h2', "LPA decisions")

    cy.contains('dt', "When your attorneys can use your LPA")
    cy.contains('dt', "Who are your attorneys")
    cy.contains('dt', "Who are your replacement attorneys")

    cy.contains('h2', "People named on the LPA")
    cy.contains('h3', "Donor")
    cy.contains('h3', "Certificate provider")
    cy.contains('h3', "Attorneys")

    cy.get('#f-checked-and-happy').check({ force: true })

    cy.contains('button', 'Confirm').click();

    cy.url().should('contain', '/lpa-details-saved');

    cy.visit('/dashboard');

    cy.contains('.govuk-body-s', 'Reference number:')
      .invoke('text')
      .then((text) => {
        const uid = text.split(':')[1].trim();
        cy.visit(`http://localhost:9001/?detail-type=notification-sent&detail=${uid}`);

        cy.contains(`"uid":"${uid}"`)
        cy.contains('"notificationId":"an-email-id"');
      });
  });

  it('does not allow checking when no changes', () => {
    cy.visit('/fixtures?redirect=/check-your-lpa&progress=peopleToNotifyAboutYourLpa');

    cy.get('#f-checked-and-happy').check({ force: true })
    cy.contains('button', 'Confirm').click();

    cy.visitLpa('/check-your-lpa');
    cy.contains('button', 'Confirm').should('not.exist');

    cy.visitLpa('/restrictions');
    cy.get('#f-restrictions').type('2');
    cy.contains('button', 'Save and continue').click();

    cy.visitLpa('/check-your-lpa');
    cy.contains('button', 'Confirm');
  });

  describe('CP acting on paper', () => {
    describe('on first check', () => {
      it('content is tailored for paper CPs, a details component is shown and nav redirects to payment', () => {
        cy.visit('/fixtures?redirect=/check-your-lpa&progress=peopleToNotifyAboutYourLpa&certificateProvider=paper');

        cy.get('label[for=f-checked-and-happy]').contains('I’ve checked this LPA and I’m happy to show it to my certificate provider, Charlie Cooper')
        cy.get('details').contains('What happens if I need to make changes later?')

        cy.get('#f-checked-and-happy').check({ force: true })
        cy.contains('button', 'Confirm').click();

        cy.url().should('contain', '/lpa-details-saved');

        cy.get('div[data-module=govuk-notification-banner]').contains('You should show your LPA to your certificate provider, Charlie Cooper.')

        cy.contains('a', 'Continue').click();

        cy.url().should('contain', '/task-list');
      })
    })

    describe('on subsequent check when LPA has not been paid for', () => {
      it('content is tailored for paper CPs, a warning component is shown and nav redirects to payment', () => {
        cy.visit('/fixtures?redirect=/check-your-lpa&progress=checkAndSendToYourCertificateProvider&certificateProvider=paper');

        cy.get('label[for=f-checked-and-happy]').contains('I’ve checked this LPA and I’m happy to show it to my certificate provider, Charlie Cooper')
        cy.get('.govuk-warning-text').contains('Once you select the confirm button, your certificate provider will be sent a text telling them you have changed your LPA.')

        cy.get('#f-checked-and-happy').check({ force: true })
        cy.contains('button', 'Confirm').click();

        cy.url().should('contain', '/lpa-details-saved');

        cy.get('div[data-module=govuk-notification-banner]').contains('We’ve saved your changes and sent a text to your certificate provider, Charlie Cooper, to tell them that your LPA is ready for review. You should show them your LPA.')

        cy.contains('a', 'Continue').click();

        cy.url().should('contain', '/task-list');
      })
    })

    describe('on subsequent check when LPA has been paid for', () => {
      it('content is tailored for paper CPs, a warning component is shown and nav redirects to dashboard', () => {
        cy.visit('/fixtures?redirect=/check-your-lpa&progress=payForTheLpa&certificateProvider=paper');

        cy.get('label[for=f-checked-and-happy]').contains('I’ve checked this LPA and I’m happy to show it to my certificate provider, Charlie Cooper')
        cy.get('.govuk-warning-text').contains('Once you select the confirm button, your certificate provider will be sent a text telling them you have changed your LPA.')

        cy.get('#f-checked-and-happy').check({ force: true })
        cy.contains('button', 'Confirm').click();

        cy.url().should('contain', '/lpa-details-saved');

        cy.get('div[data-module=govuk-notification-banner]').contains('We’ve saved your changes and sent a text to your certificate provider, Charlie Cooper, to tell them that your LPA is ready for review. You should show them your LPA.')

        cy.contains('a', 'Return to dashboard').click();

        cy.url().should('contain', '/dashboard');
      })
    })
  })

  describe('CP acting online', () => {
    describe('on first check', () => {
      it('content is tailored for online CPs, a details component is shown and nav redirects to payment', () => {
        cy.visit('/fixtures?redirect=/check-your-lpa&progress=peopleToNotifyAboutYourLpa');

        cy.get('label[for=f-checked-and-happy]').contains('I’ve checked this LPA and I’m happy for OPG to share it with my certificate provider, Charlie Cooper')
        cy.get('details').contains('What happens if I need to make changes later?')

        cy.get('#f-checked-and-happy').check({ force: true })
        cy.contains('button', 'Confirm').click();

        cy.url().should('contain', '/lpa-details-saved');

        cy.get('div[data-module=govuk-notification-banner]').contains('We’ve sent an email to your certificate provider, Charlie Cooper, to tell them what they need to do next. You should tell them to expect an email from us.')

        cy.contains('a', 'Continue').click();

        cy.url().should('contain', '/task-list');
      })
    })

    describe('on subsequent check when LPA has not been paid for', () => {
      it('content is tailored for online CPs, a warning component is shown and nav redirects to payment', () => {
        cy.visit('/fixtures?redirect=/check-your-lpa&progress=checkAndSendToYourCertificateProvider');

        cy.get('label[for=f-checked-and-happy]').contains('I’ve checked this LPA and I’m happy for OPG to share it with my certificate provider, Charlie Cooper')
        cy.get('.govuk-warning-text').contains('Once you select the confirm button, your certificate provider will be sent a text telling them you have changed your LPA.')

        cy.get('#f-checked-and-happy').check({ force: true })
        cy.contains('button', 'Confirm').click();

        cy.url().should('contain', '/lpa-details-saved');

        cy.get('div[data-module=govuk-notification-banner]').contains('We’ve saved your changes and sent a text to your certificate provider, Charlie Cooper, to tell them that they should review your LPA online.')

        cy.contains('a', 'Continue').click();

        cy.url().should('contain', '/task-list');
      })
    })

    describe('on subsequent check when LPA has been paid for', () => {
      it('content is tailored for online CPs, a warning component is shown and nav redirects to dashboard', () => {
        cy.visit('/fixtures?redirect=/check-your-lpa&progress=payForTheLpa');

        cy.get('label[for=f-checked-and-happy]').contains('I’ve checked this LPA and I’m happy for OPG to share it with my certificate provider, Charlie Cooper')
        cy.get('.govuk-warning-text').contains('Once you select the confirm button, your certificate provider will be sent a text telling them you have changed your LPA.')

        cy.get('#f-checked-and-happy').check({ force: true })
        cy.contains('button', 'Confirm').click();

        cy.url().should('contain', '/lpa-details-saved');

        cy.get('div[data-module=govuk-notification-banner]').contains('We’ve saved your changes and sent a text to your certificate provider, Charlie Cooper, to tell them that they should review your LPA online.')

        cy.contains('a', 'Return to dashboard').click();

        cy.url().should('contain', '/dashboard');
      })
    })
  })

  it("errors when not selected", () => {
    cy.visit('/fixtures?redirect=/check-your-lpa&progress=peopleToNotifyAboutYourLpa');
    cy.contains('button', 'Confirm').click();

    cy.get('.govuk-error-summary').within(() => {
      cy.contains('Select the box if you have checked your LPA and are happy to share it with your certificate provider');
    });

    cy.contains('.govuk-form-group .govuk-error-message', 'Select the box if you have checked your LPA and are happy to share it with your certificate provider');
  })
});
