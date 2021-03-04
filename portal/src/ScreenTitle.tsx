import React from "react";
import cn from "classnames";
import { Text } from "@fluentui/react";
import styles from "./ScreenTitle.module.scss";

export interface ScreenTitleProps {
  className?: string;
  children?: React.ReactNode;
}

const ScreenTitle: React.FC<ScreenTitleProps> = function ScreenTitle(
  props: ScreenTitleProps
) {
  const { className, children } = props;
  return (
    <Text as="h1" variant="xxLarge" className={cn(className, styles.title)}>
      {children}
    </Text>
  );
};

export default ScreenTitle;
