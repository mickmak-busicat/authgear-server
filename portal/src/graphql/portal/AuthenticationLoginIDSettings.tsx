import React from "react";
import produce from "immer";
import {
  Checkbox,
  Toggle,
  PrimaryButton,
  TagPicker,
  Label,
} from "@fluentui/react";

import { Context, FormattedMessage } from "@oursky/react-messageformat";

import ExtendableWidget from "../../ExtendableWidget";
import CheckboxWithContent from "../../CheckboxWithContent";
import { useCheckbox, useTagPickerWithNewTags } from "../../hook/useInput";
import {
  LoginIDKeyType,
  LoginIDKeyConfig,
  PortalAPIAppConfig,
} from "../../types";
import { setFieldIfChanged, isArrayEqualInOrder } from "../../util/misc";

import styles from "./AuthenticationLoginIDSettings.module.scss";

interface Props {
  effectiveAppConfig: PortalAPIAppConfig | null;
  rawAppConfig: PortalAPIAppConfig | null;
}

interface WidgetHeaderProps {
  enabled: boolean;
  setEnabled: (enabled: boolean) => void;
  titleId: string;
}

interface AuthenticationLoginIDSettingsState {
  usernameEnabled: boolean;
  emailEnabled: boolean;
  phoneNumberEnabled: boolean;

  excludedKeywords: string[];
  isBlockReservedUsername: boolean;
  isExcludeKeywords: boolean;
  isUsernameCaseSensitive: boolean;
  isAsciiOnly: boolean;

  isEmailCaseSensitive: boolean;
  isIgnoreDotLocal: boolean;
  isAllowPlus: boolean;
}

const switchStyle = { root: { margin: "0" } };

const WidgetHeader: React.FC<WidgetHeaderProps> = function (
  props: WidgetHeaderProps
) {
  const { titleId, enabled, setEnabled } = props;
  const onChange = React.useCallback(
    (_event, checked?: boolean) => {
      setEnabled(!!checked);
    },
    [setEnabled]
  );
  return (
    <div className={styles.widgetHeader}>
      <Toggle
        label={<FormattedMessage id={titleId} />}
        inlineLabel={true}
        styles={switchStyle}
        checked={enabled}
        onChange={onChange}
      />
    </div>
  );
};

function extractConfigFromLoginIdKeys(
  loginIdKeys: LoginIDKeyConfig[]
): { [key: string]: boolean } {
  // We consider them as enabled if they are listed as allowed login ID keys.
  const usernameEnabledConfig =
    loginIdKeys.find((key) => key.type === "username") != null;
  const emailEnabledConfig =
    loginIdKeys.find((key) => key.type === "email") != null;
  const phoneNumberEnabledConfig =
    loginIdKeys.find((key) => key.type === "phone") != null;

  return {
    usernameEnabledConfig,
    emailEnabledConfig,
    phoneNumberEnabledConfig,
  };
}

function handleStringListInput(
  stringList: string[],
  options = {
    optionEnabled: true,
    useDefaultList: false,
    defaultList: [] as string[],
  }
) {
  if (!options.optionEnabled) {
    return [];
  }
  const sanitizedList = stringList.map((item) => item.trim()).filter(Boolean);
  return options.useDefaultList
    ? [...sanitizedList, ...options.defaultList]
    : sanitizedList;
}

function setFieldIfListNonEmpty(
  map: Record<string, unknown>,
  field: string,
  list: (string | number | boolean)[]
): void {
  if (list.length === 0) {
    delete map[field];
  } else {
    map[field] = list;
  }
}

function getOrCreateLoginIdKey(
  loginIdKeys: LoginIDKeyConfig[],
  keyType: LoginIDKeyType
): LoginIDKeyConfig {
  const loginIdKey = loginIdKeys.find((key: any) => key.type === keyType);
  if (loginIdKey != null) {
    return loginIdKey;
  }
  const newLoginIdKey = { type: keyType };
  loginIdKeys.push(newLoginIdKey);
  return newLoginIdKey;
}

function setLoginIdKeyEnabled(
  loginIdKey: LoginIDKeyConfig,
  enabled: boolean,
  initialEnabled: boolean
) {
  if (enabled === initialEnabled) {
    return;
  }
  loginIdKey.verification = loginIdKey.verification ?? { enabled: false };
  loginIdKey.verification.enabled = enabled;
}

function constructStateFromAppConfig(
  appConfig: PortalAPIAppConfig | null
): AuthenticationLoginIDSettingsState {
  const loginIdKeys = appConfig?.identity?.login_id?.keys ?? [];
  const {
    usernameEnabled,
    emailEnabled,
    phoneNumberEnabled,
  } = extractConfigFromLoginIdKeys(loginIdKeys);

  // username widget
  const usernameConfig = appConfig?.identity?.login_id?.types?.username;
  const excludedKeywords = usernameConfig?.excluded_keywords ?? [];

  // email widget
  const emailConfig = appConfig?.identity?.login_id?.types?.email;

  return {
    usernameEnabled,
    emailEnabled,
    phoneNumberEnabled,

    excludedKeywords,
    isBlockReservedUsername: !!usernameConfig?.block_reserved_usernames,
    isExcludeKeywords: excludedKeywords.length > 0,
    isUsernameCaseSensitive: !!usernameConfig?.case_sensitive,
    isAsciiOnly: !!usernameConfig?.ascii_only,

    isEmailCaseSensitive: !!emailConfig?.case_sensitive,
    isIgnoreDotLocal: !!emailConfig?.ignore_dot_sign,
    isAllowPlus: !emailConfig?.block_plus_sign,
  };
}

function constructAppConfigFromState(
  rawAppConfig: PortalAPIAppConfig,
  initialScreenState: AuthenticationLoginIDSettingsState,
  screenState: AuthenticationLoginIDSettingsState
): PortalAPIAppConfig {
  const newAppConfig = produce(rawAppConfig, (draftConfig) => {
    const loginIdKeys = draftConfig.identity?.login_id?.keys ?? [];
    const loginIdUsernameKey = getOrCreateLoginIdKey(loginIdKeys, "username");
    const loginIdEmailKey = getOrCreateLoginIdKey(loginIdKeys, "email");
    const loginIdPhoneNumberKey = getOrCreateLoginIdKey(loginIdKeys, "phone");

    setLoginIdKeyEnabled(
      loginIdUsernameKey,
      screenState.usernameEnabled,
      initialScreenState.usernameEnabled
    );
    setLoginIdKeyEnabled(
      loginIdEmailKey,
      screenState.emailEnabled,
      initialScreenState.emailEnabled
    );
    setLoginIdKeyEnabled(
      loginIdPhoneNumberKey,
      screenState.phoneNumberEnabled,
      initialScreenState.phoneNumberEnabled
    );

    draftConfig.identity = draftConfig.identity ?? {};
    draftConfig.identity.login_id = draftConfig.identity.login_id ?? {};
    draftConfig.identity.login_id.types =
      draftConfig.identity.login_id.types ?? {};

    const loginIdTypes = draftConfig.identity.login_id.types;

    // username config
    loginIdTypes.username = loginIdTypes.username ?? {};
    const usernameConfig = loginIdTypes.username;

    if (
      !isArrayEqualInOrder(
        initialScreenState.excludedKeywords,
        screenState.excludedKeywords
      )
    ) {
      const excludedKeywordList = handleStringListInput(
        screenState.excludedKeywords,
        {
          optionEnabled: screenState.isExcludeKeywords,
          useDefaultList: false,
          defaultList: [],
        }
      );

      setFieldIfListNonEmpty(
        usernameConfig,
        "excluded_keywords",
        excludedKeywordList
      );
    }
    setFieldIfChanged(
      usernameConfig,
      "case_sensitive",
      initialScreenState.isUsernameCaseSensitive,
      screenState.isUsernameCaseSensitive
    );
    setFieldIfChanged(
      usernameConfig,
      "ascii_only",
      initialScreenState.isAsciiOnly,
      screenState.isAsciiOnly
    );

    // email config
    loginIdTypes.email = loginIdTypes.email ?? {};
    const emailConfig = loginIdTypes.email;

    setFieldIfChanged(
      emailConfig,
      "case_sensitive",
      initialScreenState.isEmailCaseSensitive,
      screenState.isEmailCaseSensitive
    );
    setFieldIfChanged(
      emailConfig,
      "ignore_dot_sign",
      initialScreenState.isIgnoreDotLocal,
      screenState.isIgnoreDotLocal
    );
    setFieldIfChanged(
      emailConfig,
      "block_plus_sign",
      !initialScreenState.isAllowPlus,
      !screenState.isAllowPlus
    );
  });

  return newAppConfig;
}

const AuthenticationLoginIDSettings: React.FC<Props> = function AuthenticationLoginIDSettings(
  props: Props
) {
  const { effectiveAppConfig, rawAppConfig } = props;
  const { renderToString } = React.useContext(Context);
  const initialState = React.useMemo(() => {
    return constructStateFromAppConfig(effectiveAppConfig);
  }, [effectiveAppConfig]);

  const [usernameEnabled, setUsernameEnabled] = React.useState(
    initialState.usernameEnabled
  );
  const [emailEnabled, setEmailEnabled] = React.useState(
    initialState.emailEnabled
  );
  const [phoneNumberEnabled, setPhoneNumberEnabled] = React.useState(
    initialState.phoneNumberEnabled
  );

  // username widget
  const {
    list: excludedKeywords,
    onChange: onExcludedKeywordsChange,
    defaultSelectedItems: defaultSelectedExcludedKeywords,
    onResolveSuggestions: onResolveExcludedKeywordSuggestions,
  } = useTagPickerWithNewTags(initialState.excludedKeywords);
  const {
    value: isBlockReservedUsername,
    onChange: onIsBlockReservedUsernameChange,
  } = useCheckbox(initialState.isBlockReservedUsername);
  const {
    value: isExcludeKeywords,
    onChange: onIsExcludeKeywordsChange,
  } = useCheckbox(initialState.isExcludeKeywords);
  const {
    value: isUsernameCaseSensitive,
    onChange: onIsUsernameCaseSensitiveChange,
  } = useCheckbox(initialState.isUsernameCaseSensitive);
  const { value: isAsciiOnly, onChange: onIsAsciiOnlyChange } = useCheckbox(
    initialState.isAsciiOnly
  );

  // email widget
  const {
    value: isEmailCaseSensitive,
    onChange: onIsEmailCaseSensitiveChange,
  } = useCheckbox(initialState.isEmailCaseSensitive);
  const {
    value: isIgnoreDotLocal,
    onChange: onIsIgnoreDotLocalChange,
  } = useCheckbox(initialState.isIgnoreDotLocal);
  const { value: isAllowPlus, onChange: onIsAllowPlusChange } = useCheckbox(
    initialState.isAllowPlus
  );

  // on save
  const onSaveButtonClicked = React.useCallback(() => {
    if (rawAppConfig == null) {
      return;
    }

    const screenState = {
      usernameEnabled,
      emailEnabled,
      phoneNumberEnabled,

      excludedKeywords,
      isBlockReservedUsername,
      isExcludeKeywords,
      isUsernameCaseSensitive,
      isAsciiOnly,

      isEmailCaseSensitive,
      isIgnoreDotLocal,
      isAllowPlus,
    };

    constructAppConfigFromState(rawAppConfig, initialState, screenState);
    // TODO: call mutation to save config
  }, [
    rawAppConfig,
    initialState,

    usernameEnabled,
    emailEnabled,
    phoneNumberEnabled,

    excludedKeywords,
    isBlockReservedUsername,
    isExcludeKeywords,
    isUsernameCaseSensitive,
    isAsciiOnly,

    isEmailCaseSensitive,
    isIgnoreDotLocal,
    isAllowPlus,
  ]);

  return (
    <div className={styles.root}>
      <div className={styles.widgetContainer}>
        <ExtendableWidget
          initiallyExtended={usernameEnabled}
          extendable={true}
          readOnly={!usernameEnabled}
          extendButtonAriaLabelId={"AuthenticationWidget.usernameExtend"}
          HeaderComponent={
            <WidgetHeader
              enabled={usernameEnabled}
              setEnabled={setUsernameEnabled}
              titleId={"AuthenticationWidget.usernameTitle"}
            />
          }
        >
          <div className={styles.usernameWidgetContent}>
            <Checkbox
              label={renderToString(
                "AuthenticationWidget.blockReservedUsername"
              )}
              checked={isBlockReservedUsername}
              onChange={onIsBlockReservedUsernameChange}
              className={styles.checkboxWithContent}
            />

            <CheckboxWithContent
              ariaLabel={renderToString("AuthenticationWidget.excludeKeywords")}
              checked={isExcludeKeywords}
              onChange={onIsExcludeKeywordsChange}
              className={styles.checkboxWithContent}
            >
              <Label className={styles.checkboxLabel}>
                <FormattedMessage id="AuthenticationWidget.excludeKeywords" />
              </Label>
              <TagPicker
                inputProps={{
                  "aria-label": renderToString(
                    "AuthenticationWidget.excludeKeywords"
                  ),
                }}
                className={styles.widgetInputField}
                disabled={!isExcludeKeywords}
                onChange={onExcludedKeywordsChange}
                defaultSelectedItems={defaultSelectedExcludedKeywords}
                onResolveSuggestions={onResolveExcludedKeywordSuggestions}
              />
            </CheckboxWithContent>

            <Checkbox
              label={renderToString("AuthenticationWidget.caseSensitive")}
              className={styles.widgetCheckbox}
              checked={isUsernameCaseSensitive}
              onChange={onIsUsernameCaseSensitiveChange}
            />

            <Checkbox
              label={renderToString("AuthenticationWidget.asciiOnly")}
              className={styles.widgetCheckbox}
              checked={isAsciiOnly}
              onChange={onIsAsciiOnlyChange}
            />
          </div>
        </ExtendableWidget>
      </div>

      <div className={styles.widgetContainer}>
        <ExtendableWidget
          initiallyExtended={emailEnabled}
          extendable={true}
          readOnly={!emailEnabled}
          extendButtonAriaLabelId={"AuthenticationWidget.emailAddressExtend"}
          HeaderComponent={
            <WidgetHeader
              enabled={emailEnabled}
              setEnabled={setEmailEnabled}
              titleId={"AuthenticationWidget.emailAddressTitle"}
            />
          }
        >
          <Checkbox
            label={renderToString("AuthenticationWidget.caseSensitive")}
            className={styles.widgetCheckbox}
            checked={isEmailCaseSensitive}
            onChange={onIsEmailCaseSensitiveChange}
          />

          <Checkbox
            label={renderToString("AuthenticationWidget.ignoreDotLocal")}
            className={styles.widgetCheckbox}
            checked={isIgnoreDotLocal}
            onChange={onIsIgnoreDotLocalChange}
          />

          <Checkbox
            label={renderToString("AuthenticationWidget.allowPlus")}
            className={styles.widgetCheckbox}
            checked={isAllowPlus}
            onChange={onIsAllowPlusChange}
          />
        </ExtendableWidget>
      </div>

      <div className={styles.widgetContainer}>
        <ExtendableWidget
          initiallyExtended={phoneNumberEnabled}
          extendable={true}
          readOnly={!phoneNumberEnabled}
          extendButtonAriaLabelId={"AuthenticationWidget.phoneNumberExtend"}
          HeaderComponent={
            <WidgetHeader
              enabled={phoneNumberEnabled}
              setEnabled={setPhoneNumberEnabled}
              titleId={"AuthenticationWidget.phoneNumberTitle"}
            />
          }
        >
          <div>TODO: To be implemented</div>
        </ExtendableWidget>
      </div>
      <div className={styles.saveButtonContainer}>
        <PrimaryButton onClick={onSaveButtonClicked}>
          <FormattedMessage id="save" />
        </PrimaryButton>
      </div>
    </div>
  );
};

export default AuthenticationLoginIDSettings;
