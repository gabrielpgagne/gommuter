import streamlit as st
import pandas as pd
import os


def load_commute_time(file):
    try:
        df = pd.read_csv(file, names=["datetime", "commute_time"])
        df_dt = pd.to_datetime(df["datetime"])
        df["hour_min"] = df_dt.dt.strftime("%H:%M")
        df_weekday_nbr = df_dt.dt.weekday + 1
        df_weekday_name = df_dt.dt.day_name()
        df["weekday"] = df_weekday_nbr.astype(str) + " - " + df_weekday_name
    except Exception as e:
        print(e)
        df = pd.DataFrame({"hour_min": [], "commute_time": []})
    return df


def get_average_commute_time(df: pd.DataFrame):
    avg_time = df.groupby("hour_min")["commute_time"].mean()
    avg_time = avg_time.reset_index()
    return avg_time


def get_average_commute_time_daywise(df: pd.DataFrame):
    avg_time_daywise = df.groupby(["weekday", "hour_min"])["commute_time"].mean()
    avg_time_daywise = avg_time_daywise.reset_index()
    return avg_time_daywise


def get_data_ids():
    ids = []
    for f in os.listdir("data"):
        id = f.split("-")[-1].replace(".csv", "")
        if id.isdigit():
            ids.append(id)
    ids = list(set(ids))
    ids.sort()
    return ids


def make_tab(id):
    col1, col2 = st.columns(2, border=True)
    with col1:
        st.header("Commute time - to")
        df = load_commute_time(f"data/to-{id}.csv")

        st.subheader("Average commute time")
        adf = get_average_commute_time(df)
        adf["+/-"] = df.groupby("hour_min")["commute_time"].std().values
        st.bar_chart(
            adf,
            x="hour_min",
            y="commute_time",
            x_label="Departure time (HH:MM)",
            y_label="Average commute time (min)",
            color="+/-",
        )

        st.subheader("Average day-wise commute time")
        adf = get_average_commute_time_daywise(df)
        st.bar_chart(
            adf,
            x="hour_min",
            y="commute_time",
            x_label="Departure time (HH:MM)",
            y_label="Average commute time (min)",
            stack=False,
            color="weekday",
        )
    if st.checkbox(f"Show 'to-{id}' data"):
        st.dataframe(df)

    with col2:
        st.header("Commute time - from")
        df = load_commute_time(f"data/from-{id}.csv")

        st.subheader("Average commute time")
        adf = get_average_commute_time(df)
        adf["+/-"] = df.groupby("hour_min")["commute_time"].std().values
        st.bar_chart(
            adf,
            x="hour_min",
            y="commute_time",
            x_label="Departure time (HH:MM)",
            y_label="Average commute time (min)",
            color="+/-",
        )

        st.subheader("Average day-wise commute time")
        adf = get_average_commute_time_daywise(df)
        st.bar_chart(
            adf,
            x="hour_min",
            y="commute_time",
            x_label="Departure time (HH:MM)",
            y_label="Average commute time (min)",
            stack=False,
            color="weekday",
        )

    if st.checkbox(f"Show 'from-{id}' data"):
        st.dataframe(df)


# Run the app
if __name__ == "__main__":
    st.set_page_config(page_title="Commute time dashboard", layout="wide")

    ids = get_data_ids()
    tabs = st.tabs([str(i) for i in ids])
    for i, id in enumerate(ids):
        with tabs[i]:
            st.header(f"Itinerary {id}")
            make_tab(id)
