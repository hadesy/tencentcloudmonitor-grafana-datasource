import zh_CN from './zh_CN';
import en_US from './en_US';

let locale_language = 'zh-CN'

export enum Language {
    Chinese = 'zh-CN',
    English = 'en-US',
}

export const t = (key: string) => {
    if (locale_language === Language.Chinese) {
        // @ts-ignore
        return zh_CN[key]
    }
    // @ts-ignore
    return en_US[key];
}

export const setLanguage = (language: Language) => {
    locale_language = language;
}

export const getLanguage = () => locale_language;
